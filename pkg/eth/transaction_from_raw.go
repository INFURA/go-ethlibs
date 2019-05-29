package eth

import (
	"github.com/pkg/errors"

	"github.com/INFURA/ethereum-interaction/pkg/rlp"
)

// FromRaw populates a Transaction fields from the RLP-encoded raw transaction input string.
// Currently only supports signed Transactions,
func (t *Transaction) FromRaw(input string) error {
	// Code is heavily inspired by ethers.js utils.transaction.parse:
	// https://github.com/ethers-io/ethers.js/blob/master/utils/transaction.js#L90
	// Copyright (c) 2017 Richard Moore

	// Decode the input string as an rlp.Value
	decoded, err := rlp.From(input)
	if err != nil {
		return errors.Wrap(err, "could not RLP decode raw input")
	}

	// Signed transactions should be an RLP list with nine fields
	items := decoded.List
	switch len(items) {
	case 0:
		return errors.New("raw input decoded to non-list or empty list")
	case 6:
		// TODO: this code does not (yet) support unsigned transactions
		return errors.New("unsigned transactions not supported")
	case 9:
		// expected
		break
	default:
		return errors.Errorf("unexpected decoded list size %d", len(decoded.List))
	}

	// list item 0 is the Nonce
	nonce, err := NewQuantityFromRLP(items[0])
	if err != nil {
		return errors.Wrap(err, "could not parse transaction nonce")
	}

	// list item 1 is the GasPrice
	gasPrice, err := NewQuantityFromRLP(items[1])
	if err != nil {
		return errors.Wrap(err, "could not parse transaction gasPrice")
	}

	// list item 2 is the GasLimit
	gasLimit, err := NewQuantityFromRLP(items[2])
	if err != nil {
		return errors.Wrap(err, "could not parse transaction gasLimit")
	}

	// list item 3 is the To address, which may be nil (0x) for contract creation
	var to *Address
	if items[3].String != "0x" {
		t, err := NewAddress(items[3].String)
		if err != nil {
			return errors.Wrap(err, "could not parse transaction to address")
		}

		to = t
	}

	// list item 4 is the ETH Value
	value, err := NewQuantityFromRLP(items[4])
	if err != nil {
		return errors.Wrap(err, "could not parse transaction value")
	}

	// list item 5 is the transaction payload data/contract method input
	data, err := NewData(items[5].String)
	if err != nil {
		return errors.Wrap(err, "could not parse transaction data")
	}

	// TODO: unsigned transactions do not include these next fields

	// list item 6 is the signature's "V"
	v, err := NewQuantityFromRLP(items[6])
	if err != nil {
		return errors.Wrap(err, "could not parse transaction v field")
	}

	// list item 7 is the signature's "R"
	r, err := NewQuantityFromRLP(items[7])
	if err != nil {
		return errors.Wrap(err, "could not parse transaction r field")
	}

	// list item 8 is the signature's "S"
	s, err := NewQuantityFromRLP(items[8])
	if err != nil {
		return errors.Wrap(err, "could not parse transaction s field")
	}

	// Get the Chain Id from the V value, see EIP155:
	//   https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md#list-of-chain-ids
	chainId := (v.Int64() - 35) / 2
	if chainId < 0 {
		chainId = 0
	}

	// Sender Address recovery.  Since the transaction does not include the senders address we have to
	// recover it by deriving the public key used to sign the transaction, and then converting said
	// public key into an Ethereuem address.
	//
	// To do so, the first step is to compute the original hash that was signed with
	// the R/S/V values.  Which for EIP155 transactions is the RLP encoding of the list:
	//   [ Nonce, GasPrice, GasLimit, To, Value, Input, ChainId, 0, 0 ]
	// While for non-EIP155 transactions (where ChainId is zero), only the first six fields are hashed:
	//   [ Nonce, GasPrice, GasLimit, To, Value, Input ]
	raw := rlp.Value{
		List: []rlp.Value{
			items[0], items[1], items[2], items[3], items[4], items[5],
		},
	}

	// Compute the "recovery" V used when reversing the signing, which for non-EIP155 transactions
	// is just V - 27.
	recovery := v.Int64() - 27

	if chainId != 0 {
		// as mentioned above, append [ ChainId, 0, 0 ] to the hash input
		raw.List = append(raw.List,
			QuantityFromInt64(chainId).RLP(),
			rlp.Value{String: "0x"},
			rlp.Value{String: "0x"},
		)

		// And subtract out the chainId to get a proper recovery value
		recovery -= (chainId * 2) + 8
	}

	recoveryV := QuantityFromInt64(recovery)

	// We can now compute the input hash for address recovery
	rh, err := raw.Hash()
	if err != nil {
		return errors.Wrap(err, "could not hash raw transaction")
	}
	recoverHash, err := NewHash(rh)
	if err != nil {
		return errors.Wrap(err, "invalid raw transaction hash")
	}

	// And now we can pass these values into ECRecover, which returns the sender address
	sender, err := ECRecover(recoverHash, r, s, &recoveryV)
	if err != nil {
		return errors.Wrap(err, "could not recover from address")
	}

	// And finally, we need to compute the transaction's Hash...
	dh, err := decoded.Hash()
	if err != nil {
		return errors.Wrap(err, "could not hash decoded transaction")
	}

	txHash, err := NewHash(dh)
	if err != nil {
		return errors.Wrap(err, "invalid decoded hash")
	}

	// ... and fill in all our fields with the decoded values
	t.Nonce = *nonce
	t.GasPrice = *gasPrice
	t.Gas = *gasLimit
	t.To = to
	t.Value = *value
	t.Input = *data
	t.V = *v
	t.R = *r
	t.S = *s
	t.Hash = *txHash
	t.From = *sender
	return nil
}
