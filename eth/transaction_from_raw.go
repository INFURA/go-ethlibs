package eth

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/INFURA/go-ethlibs/rlp"
)

// FromRaw populates a Transaction fields from the raw transaction data supplied as a hexadecimal encoded string.
// For pre-EIP-2718 legacy transactions the input string is an RLP-encoded list, for transaction types defined
// after EIP-2718 the payload format depends on the transaction type included as the first byte.
// Unsigned transactions where R, S, and V are zero are not currently supported.
func (t *Transaction) FromRaw(input string) error {
	// Code is heavily inspired by ethers.js utils.transaction.parse:
	// https://github.com/ethers-io/ethers.js/blob/master/utils/transaction.js#L90
	// Copyright (c) 2017 Richard Moore

	var (
		chainId    Quantity
		nonce      Quantity
		gasPrice   Quantity
		gasLimit   Quantity
		to         *Address
		value      Quantity
		data       Data
		v          Quantity
		r          Quantity
		s          Quantity
		accessList AccessList
	)

	switch {
	case strings.HasPrefix(input, "0x01"):
		// EIP-2930 transaction
		payload := "0x" + input[4:]
		if err := rlpDecodeList(payload, &chainId, &nonce, &gasPrice, &gasLimit, &to, &value, &data, &accessList, &v, &r, &s); err != nil {
			return errors.Wrap(err, "could not decode RLP components")
		}

		t.Type = MustQuantity("0x1")
		t.Nonce = nonce
		t.GasPrice = gasPrice
		t.Gas = gasLimit
		t.To = to
		t.Value = value
		t.Input = data
		t.AccessList = &accessList
		t.V = v
		t.R = r
		t.S = s
		t.ChainId = &chainId

		signingHash, err := t.SigningHash(chainId)
		if err != nil {
			return err
		}

		if r.Int64() == 0 && s.Int64() == 0 {
			return errors.New("unsigned transactions not supported")
		}

		sender, err := ECRecover(signingHash, &r, &s, &v)
		if err != nil {
			return err
		}

		raw, err := t.RawRepresentation(chainId)
		if err != nil {
			return err
		}

		t.Hash = raw.Hash()
		t.From = *sender
		return nil
	default:
		// Legacy Transaction
		// Decode the input string as an rlp.Value
		if err := rlpDecodeList(input, &nonce, &gasPrice, &gasLimit, &to, &value, &data, &v, &r, &s); err != nil {
			return errors.Wrap(err, "could not decode RLP components")
		}

		// ... and fill in all our fields with the decoded values
		t.Nonce = nonce
		t.GasPrice = gasPrice
		t.Gas = gasLimit
		t.To = to
		t.Value = value
		t.Input = data
		t.V = v
		t.R = r
		t.S = s

		// Pull out chainId and recoveryV from EIP-155 packed V
		_chain := (v.Int64() - 35) / 2
		if _chain < 0 {
			_chain = 0
		}

		_recovery := v.Int64() - 27
		if _chain != 0 {
			// And subtract out the chainId to get a proper recovery value
			_recovery -= (_chain * 2) + 8
		}

		recoveryV := QuantityFromInt64(_recovery)
		chainId = QuantityFromInt64(_chain)

		signingHash, err := t.SigningHash(chainId)
		if err != nil {
			return err
		}

		sender, err := ECRecover(signingHash, &r, &s, &recoveryV)
		if err != nil {
			return err
		}

		raw, err := t.RawRepresentation(chainId)
		if err != nil {
			return err
		}

		t.Hash = raw.Hash()
		t.From = *sender
		return nil
	}
}

func rlpDecodeList(input string, receivers ...interface{}) error {
	decoded, err := rlp.From(input)
	if err != nil {
		return err
	}

	if len(decoded.List) < len(receivers) {
		return errors.Errorf("expected %d items but only received %d", len(receivers), len(decoded.List))
	}

	for i := range receivers {
		value := decoded.List[i]
		switch receiver := receivers[i].(type) {
		case *Quantity:
			q, err := NewQuantityFromRLP(value)
			if err != nil {
				return errors.Wrapf(err, "could not decode list item %d to Quantity", i)
			}
			*receiver = *q
		case **Address:
			if value.String == "0x" {
				*receiver = nil
			} else {
				a, err := NewAddress(value.String)
				if err != nil {
					return errors.Wrapf(err, "could not decode list item %d to Address", i)
				}
				*receiver = a
			}
		case *Data:
			d, err := NewData(value.String)
			if err != nil {
				return errors.Wrapf(err, "could not decode list item %d to Data", i)
			}
			*receiver = *d
		case *rlp.Value:
			*receiver = value
		case *AccessList:
			accessList, err := NewAccessListFromRLP(value)
			if err != nil {
				return errors.Wrapf(err, "could not decode list item %d to AccessList", i)
			}
			*receiver = accessList
		default:
			return errors.Errorf("unsupported decode receiver %s", reflect.TypeOf(receiver).String())
		}
	}

	return nil
}
