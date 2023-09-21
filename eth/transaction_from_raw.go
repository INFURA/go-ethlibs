package eth

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"

	"github.com/INFURA/go-ethlibs/rlp"
)

// FromRaw populates a Transaction's fields from the raw transaction data supplied as a hexadecimal encoded string.
// For pre-EIP-2718 legacy transactions the input string is an RLP-encoded list, for transaction types defined
// after EIP-2718 the payload format depends on the transaction type included as the first byte.
// Unsigned transactions where R, S, and V are zero are not currently supported.
func (t *Transaction) FromRaw(input string) error {
	// Code was originally heavily inspired by ethers.js v4 utils.transaction.parse:
	// https://github.com/ethers-io/ethers.js/blob/v4-legacy/utils/transaction.js#L90
	// Copyright (c) 2017 Richard Moore
	//
	// However it's since been somewhat extensively rewritten to support EIP-2718 and -2930

	var (
		chainId              Quantity
		nonce                Quantity
		gasPrice             Quantity
		gasLimit             Quantity
		maxPriorityFeePerGas Quantity
		maxFeePerGas         Quantity
		to                   *Address
		value                Quantity
		data                 Data
		v                    Quantity
		r                    Quantity
		s                    Quantity
		accessList           AccessList
		maxFeePerBlobGas     Quantity
		blobVersionedHashes  []Hash
	)

	if !strings.HasPrefix(input, "0x") {
		return errors.New("input must start with 0x")
	}

	if len(input) < 4 {
		return errors.New("not enough input to decode")
	}

	var firstByte byte
	if prefix, err := NewData(input[:4]); err != nil {
		return errors.Wrap(err, "could not inspect transaction prefix")
	} else {
		firstByte = prefix.Bytes()[0]
	}

	switch {
	case firstByte == byte(TransactionTypeAccessList):
		// EIP-2930 transaction
		payload := "0x" + input[4:]
		if err := rlpDecodeList(payload, &chainId, &nonce, &gasPrice, &gasLimit, &to, &value, &data, &accessList, &v, &r, &s); err != nil {
			return errors.Wrap(err, "could not decode RLP components")
		}

		if r.Int64() == 0 && s.Int64() == 0 {
			return errors.New("unsigned transactions not supported")
		}

		t.Type = OptionalQuantityFromInt(int(firstByte))
		t.Nonce = nonce
		t.GasPrice = &gasPrice
		t.Gas = gasLimit
		t.To = to
		t.Value = value
		t.Input = data
		t.AccessList = &accessList
		t.V = v
		t.YParity = &v
		t.R = r
		t.S = s
		t.ChainId = &chainId

		signingHash, err := t.SigningHash(chainId)
		if err != nil {
			return err
		}

		signature, err := NewEIP2718Signature(chainId, r, s, v)
		if err != nil {
			return err
		}

		sender, err := signature.Recover(signingHash)
		if err != nil {
			return err
		}

		raw, err := t.RawRepresentation()
		if err != nil {
			return err
		}

		t.Hash = raw.Hash()
		t.From = *sender
		return nil
	case firstByte == byte(TransactionTypeDynamicFee):
		// EIP-1559 transaction
		payload := "0x" + input[4:]
		// 0x02 || rlp([chainId, nonce, maxPriorityFeePerGas, maxFeePerGas, gasLimit, to, value, data, access_list, signatureYParity, signatureR, signatureS])
		if err := rlpDecodeList(payload, &chainId, &nonce, &maxPriorityFeePerGas, &maxFeePerGas, &gasLimit, &to, &value, &data, &accessList, &v, &r, &s); err != nil {
			return errors.Wrap(err, "could not decode RLP components")
		}

		if r.Int64() == 0 && s.Int64() == 0 {
			return errors.New("unsigned transactions not supported")
		}

		t.Type = OptionalQuantityFromInt(int(firstByte))
		t.Nonce = nonce
		t.MaxPriorityFeePerGas = &maxPriorityFeePerGas
		t.MaxFeePerGas = &maxFeePerGas
		t.Gas = gasLimit
		t.To = to
		t.Value = value
		t.Input = data
		t.AccessList = &accessList
		t.V = v
		t.YParity = &v
		t.R = r
		t.S = s
		t.ChainId = &chainId

		signingHash, err := t.SigningHash(chainId)
		if err != nil {
			return err
		}

		signature, err := NewEIP2718Signature(chainId, r, s, v)
		if err != nil {
			return err
		}

		sender, err := signature.Recover(signingHash)
		if err != nil {
			return err
		}

		raw, err := t.RawRepresentation()
		if err != nil {
			return err
		}

		t.Hash = raw.Hash()
		t.From = *sender
		return nil
	case firstByte == byte(TransactionTypeBlob):
		// EIP-4844 transaction
		var (
			body        rlp.Value
			blobs       []Data
			commitments []Data
			proofs      []Data
			hasBlobs    bool = false
		)
		// The raw tx can be a full "Network Representation" tx of the form:
		// 0x03 || rlp([tx_payload_body, blobs, commitments, proofs])
		//
		// Or just the tx payload body:
		// 0x03 || rlp([chain_id, nonce, max_priority_fee_per_gas, max_fee_per_gas, gas_limit, to, value, data, access_list, max_fee_per_blob_gas, blob_versioned_hashes, y_parity, r, s])
		payload := "0x" + input[4:]
		decoded, err := rlp.From(payload)

		switch len(decoded.List) {
		case 14:
			body = *decoded
			hasBlobs = false
		case 4:
			// TransactionPayloadBody is itself an RLP list of:
			// [chain_id, nonce, max_priority_fee_per_gas, max_fee_per_gas, gas_limit, to, value, data, access_list, max_fee_per_blob_gas, blob_versioned_hashes, y_parity, r, s]
			if err := rlpDecodeList(decoded, &body, &blobs, &commitments, &proofs); err != nil {
				return err
			}
			hasBlobs = true
		default:
			return errors.New("blob transaction invalid tx RLP length")
		}

		if err := rlpDecodeList(&body, &chainId, &nonce, &maxPriorityFeePerGas, &maxFeePerGas, &gasLimit, &to, &value, &data, &accessList, &maxFeePerBlobGas, &blobVersionedHashes, &v, &r, &s); err != nil {
			return errors.Wrap(err, "could not decode RLP components")
		}

		if hasBlobs && (len(blobVersionedHashes) != len(blobs) || len(blobs) != len(commitments) || len(commitments) != len(proofs)) {
			return errors.New("mismatched blob field counts")
		}

		// TODO: at this point we could verify these two constraints
		//
		// - blobVersionedHashes[i] = "0x01" + sha256(commitments[i])[4:]
		// - the KZG commitments match the corresponding blobs and proofs
		//
		// However, this all requires additionally implementing a sha256 method
		// for eth.Data and the use of a KZG proof framework, both of which
		// feel outside the scope of this package, especially considering
		// that these fields are not exposed at the JSONRPC layer which is
		// our primary focus.
		//
		// In pseudocode this all would look something like:
		//
		//  for i := range blobVersionedHashes {
		//	  blobHash := commitments[i].Sha256()
		//	  versionedHash := "0x01" + blobHash[4:]
		//	  if blobVersionedHashes[i] != versionedHash {
		//	    return errors.New("incorrect blob versioned hash")
		//	  }
		//	  if err := kzg_verify_proofs(commitments[i], blobs[i], proofs[i]); err != nil {
		//	    return err
		//	  }
		//  }
		//

		if r.Int64() == 0 && s.Int64() == 0 {
			return errors.New("unsigned transactions not supported")
		}

		t.Type = OptionalQuantityFromInt(int(firstByte))
		t.ChainId = &chainId
		t.Nonce = nonce
		t.MaxPriorityFeePerGas = &maxPriorityFeePerGas
		t.MaxFeePerGas = &maxFeePerGas
		t.Gas = gasLimit
		t.To = to
		t.Value = value
		t.Input = data
		t.AccessList = &accessList
		t.MaxFeePerBlobGas = &maxFeePerBlobGas
		t.BlobVersionedHashes = blobVersionedHashes
		t.V = v
		t.YParity = &v
		t.R = r
		t.S = s

		if hasBlobs {
			t.BlobBundle = &BlobsBundleV1{
				Blobs:       blobs,
				Commitments: commitments,
				Proofs:      proofs,
			}
		} else {
			t.BlobBundle = nil
		}

		signature, err := NewEIP2718Signature(chainId, r, s, v)
		if err != nil {
			return err
		}

		signingHash, err := t.SigningHash(signature.chainId)
		if err != nil {
			return err
		}

		sender, err := signature.Recover(signingHash)
		if err != nil {
			return err
		}

		raw, err := t.RawRepresentation()
		if err != nil {
			return err
		}

		t.Hash = raw.Hash()
		t.From = *sender
		return nil

	case firstByte > 0x7f:
		// In EIP-2718 types larger than 0x7f are reserved since they potentially conflict with legacy RLP encoded
		// transactions.  As such we can attempt to decode any such transactions as legacy format and attempt to
		// decode the input string as an rlp.Value
		if err := rlpDecodeList(input, &nonce, &gasPrice, &gasLimit, &to, &value, &data, &v, &r, &s); err != nil {
			return errors.Wrap(err, "could not decode RLP components")
		}

		if r.Int64() == 0 && s.Int64() == 0 {
			return errors.New("unsigned transactions not supported")
		}

		// ... and fill in all our fields with the decoded values
		t.Nonce = nonce
		t.GasPrice = &gasPrice
		t.Gas = gasLimit
		t.To = to
		t.Value = value
		t.Input = data
		t.V = v
		t.YParity = nil
		t.R = r
		t.S = s

		signature, err := NewEIP155Signature(r, s, v)
		if err != nil {
			return err
		}

		signingHash, err := t.SigningHash(signature.chainId)
		if err != nil {
			return err
		}

		sender, err := signature.Recover(signingHash)
		if err != nil {
			return err
		}

		raw, err := t.RawRepresentation()
		if err != nil {
			return err
		}

		t.Hash = raw.Hash()
		t.From = *sender
		return nil
	default:
		return errors.New("unsupported transaction type")
	}
}

// rlpDecodeList decodes an RLP list into the passed in receivers.  Currently only the receiver types needed for
// legacy and EIP-2930 transactions are implemented, new receivers can easily be added in the for loop.
//
// input is either a string or pointer to an rlp.Value, if it's a string then it's assumed to be RLP encoded and is decoded first
//
// Note that when calling this function, the receivers MUST be pointers never values, and for "optional" receivers
// such as Address a pointer to a pointer must be passed.  For example:
//
//    var (
//      addr  *eth.Address
//      nonce eth.Quantity
//    )
//    err := rlpDecodeList(payload, &addr, &nonce)
//
// TODO: Consider making this function public once all receiver types in the eth package are supported.
func rlpDecodeList(input interface{}, receivers ...interface{}) error {
	var decoded *rlp.Value
	switch i := input.(type) {
	case string:
		if d, err := rlp.From(i); err != nil {
			return err
		} else {
			decoded = d
		}
	case *rlp.Value:
		decoded = i
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
		case *[]Data:
			*receiver = make([]Data, len(value.List))
			for j := range value.List {
				d, err := NewData(value.List[j].String)
				if err != nil {
					return errors.Wrapf(err, "could not decode list item %d %d to Data", i, j)
				}
				(*receiver)[j] = *d
			}
		case *[]Data32:
			*receiver = make([]Data32, len(value.List))
			for j := range value.List {
				d, err := NewData32(value.List[j].String)
				if err != nil {
					return errors.Wrapf(err, "could not decode list item %d %d to Data", i, j)
				}
				(*receiver)[j] = *d
			}
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
