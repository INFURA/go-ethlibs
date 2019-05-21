package eth

import (
	"strconv"

	"github.com/pkg/errors"

	"github.com/INFURA/ethereum-interaction/pkg/rlp"
)

func (t *Transaction) FromRaw(input string) error {
	// Code is heavily inspired by ethers.js utils.transaction.parse:
	// https://github.com/ethers-io/ethers.js/blob/master/utils/transaction.js#L90
	// Copyright (c) 2017 Richard Moore

	decoded, err := rlp.From(input)
	if err != nil {
		return errors.Wrap(err, "could not RLP decode raw input")
	}

	switch len(decoded.List) {
	case 0:
		return errors.New("raw input decoded to non-list or empty list")
	case 6:
		return errors.New("unsigned transactions not supported")
	case 9:
		// good
		break
	default:
		return errors.Errorf("unexpected decoded list size %d", len(decoded.List))
	}

	dh, err := decoded.Hash()
	if err != nil {
		return errors.Wrap(err, "could not hash decoded transaction")
	}

	txHash, err := NewHash(dh)
	if err != nil {
		return errors.Wrap(err, "invalid decoded hash")
	}

	items := decoded.List
	nonce, err := NewQuantity(items[0].String)
	if err != nil {
		return errors.Wrap(err, "could not parse transaction nonce")
	}

	gasPrice, err := NewQuantity(items[1].String)
	if err != nil {
		return errors.Wrap(err, "could not parse transaction gasPrice")
	}

	gasLimit, err := NewQuantity(items[2].String)
	if err != nil {
		return errors.Wrap(err, "could not parse transaction gasLimit")
	}

	to, err := NewAddress(items[3].String)
	if err != nil {
		return errors.Wrap(err, "could not parse transaction to address")
	}

	value, err := NewQuantity(items[4].String)
	if err != nil {
		return errors.Wrap(err, "could not parse transaction value")
	}

	data, err := NewData(items[5].String)
	if err != nil {
		return errors.Wrap(err, "could not parse transaction data")
	}

	// TODO: unsigned transactions end here

	v, err := NewQuantity(items[6].String)
	if err != nil {
		return errors.Wrap(err, "could not parse transaction v field")
	}

	r, err := NewQuantity(items[7].String)
	if err != nil {
		return errors.Wrap(err, "could not parse transaction r field")
	}

	s, err := NewQuantity(items[8].String)
	if err != nil {
		return errors.Wrap(err, "could not parse transaction s field")
	}

	chainId := (v.Int64() - 35) / 2
	if chainId < 0 {
		chainId = 0
	}

	recovery := v.Int64() - 27

	raw := rlp.Value{
		List: []rlp.Value{
			items[0], items[1], items[2], items[3], items[4], items[5],
		},
	}

	if chainId != 0 {
		raw.List = append(raw.List,
			rlp.Value{String: "0x" + strconv.FormatInt(chainId, 16)},
			rlp.Value{String: MustData("0x00").String()},
			rlp.Value{String: MustData("0x00").String()},
		)
		recovery -= (chainId * 2) + 8
	}

	rh, err := raw.Hash()
	if err != nil {
		return errors.Wrap(err, "could not hash raw transaction")
	}
	recoverHash, err := NewHash(rh)
	if err != nil {
		return errors.Wrap(err, "invalid raw transaction hash")
	}

	recoverV := QuantityFromInt64(recovery)
	from, err := ECRecover(recoverHash, r, s, &recoverV)
	if err != nil {
		return errors.Wrap(err, "could not recover from address")
	}

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
	t.From = *from
	return nil
}
