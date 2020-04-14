package eth

import (
	"encoding/hex"
	"errors"

	"github.com/INFURA/go-ethlibs/rlp"
)

var ErrInsufficientParams = errors.New("transaction is missing values")
var ErrFatalCrytpo = errors.New("unable to sign Tx with private key")

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

	rawTx, err := t.serialize(chainId, false)
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
	//privKey, pubKey := secp256k1.PrivKeyFromBytes(secp256k1.S256(), pKey)
	signature, err := ECSign(hash, pKey, chainID)
	if err != nil {
		return "", err
	}

	t.V = signature.V
	t.R = signature.R
	t.S = signature.S

	signed, err := t.serialize(chainId, true)
	encoded, err := signed.Encode()

	return encoded, err
}

func (t *Transaction) serialize(chainId uint64, signature bool) (rlp.Value, error) {

	var list []rlp.Value
	list = append(list, t.Nonce.RLP())
	list = append(list, t.GasPrice.RLP())
	list = append(list, t.Gas.RLP())
	list = append(list, t.To.RLP())
	list = append(list, t.Value.RLP())
	list = append(list, rlp.Value{String: t.Input.String()})

	if !signature {
		empty := rlp.Value{String: ""}
		if chainId != 0 {
			list = append(list, QuantityFromUInt64(chainId).RLP())
			r, err := rlp.From("0x")
			if err != nil {
				return empty, err
			}
			list = append(list, *r)
			s, err := rlp.From("0x")
			if err != nil {
				return empty, err
			}
			list = append(list, *s)
		}
		rawTx := rlp.Value{List: list}
		return rawTx, nil
	} else {
		list = append(list, t.V.RLP())
		list = append(list, t.R.RLP())
		list = append(list, t.S.RLP())
		rawTx := rlp.Value{List: list}
		return rawTx, nil
	}
}

func (t *Transaction) check() bool {
	// unsure if/how best to check other values
	if t.To == nil {
		return false
	}
	return true
}
