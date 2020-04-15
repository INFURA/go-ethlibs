package eth

import (
	"encoding/hex"
	"errors"

	"github.com/INFURA/go-ethlibs/rlp"
)

var ErrInsufficientParams = errors.New("transaction is missing values")

// Looks like there have been multiple attempts to get the Koblitz curve (secp256k1) supported in golang
//
// https://github.com/golang/go/pull/26873 <-- rejected
// https://github.com/golang/go/issues/26776 <-- rejected
// using "github.com/btcsuite/btcd/btcec"
// other alternative is to import the C library but want to avoid that if possible

func (t *Transaction) Sign(privateKey string, chainId uint64) (string, error) {

	if !t.check() {
		return "", ErrInsufficientParams
	}

	pKey, err := hex.DecodeString(privateKey)
	if err != nil {
		return "", err
	}

	rawTx, err := t.serialize(chainId)
	if err != nil {
		return "", err
	}

	h, err := rawTx.Hash()
	if err != nil {
		return "", err
	}

	hash, err := NewHash(h)
	if err != nil {
		return "", err
	}

	chainID := QuantityFromUInt64(chainId)
	signature, err := ECSign(hash, pKey, chainID)
	if err != nil {
		return "", err
	}

	t.V = signature.V
	t.R = signature.R
	t.S = signature.S

	return t.RLP().Encode()
}

func (t *Transaction) RLP() rlp.Value {
	base := t.serializeCommon()
	base = append(base, t.V.RLP())
	base = append(base, t.R.RLP())
	base = append(base, t.S.RLP())
	rawTx := rlp.Value{List: base}
	return rawTx
}

func (t *Transaction) serializeCommon() []rlp.Value {
	var list []rlp.Value
	list = append(list, t.Nonce.RLP())
	list = append(list, t.GasPrice.RLP())
	list = append(list, t.Gas.RLP())
	list = append(list, t.To.RLP())
	list = append(list, t.Value.RLP())
	list = append(list, rlp.Value{String: t.Input.String()})
	return list
}

func (t *Transaction) serialize(chainID uint64) (rlp.Value, error) {
	base := t.serializeCommon()
	empty := rlp.Value{String: ""}
	if chainID != 0 {
		base = append(base, QuantityFromUInt64(chainID).RLP())
		r, err := rlp.From("0x")
		if err != nil {
			return empty, err
		}
		base = append(base, *r)
		s, err := rlp.From("0x")
		if err != nil {
			return empty, err
		}
		base = append(base, *s)
	}
	rawTx := rlp.Value{List: base}
	return rawTx, nil
}

func (t *Transaction) check() bool {
	// unsure if/how best to check other values
	if t.To == nil {
		return false
	}
	return true
}
