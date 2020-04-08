package eth

import (
	"encoding/hex"
	"errors"
	"log"

	secp256k1 "github.com/btcsuite/btcd/btcec"

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

	rawTx, err := t.serialize(chainId,false)
	if err != nil {
		return "", err
	}

	privKey, pubKey := secp256k1.PrivKeyFromBytes(secp256k1.S256(), pKey)
	signature, err := privKey.Sign([]byte(rawTx))
	if err != nil {
		return "", err
	}

	verified := signature.Verify([]byte(rawTx), pubKey)
	if !verified {
		panic(ErrFatalCrytpo)
	}

	// extract signature (r,s,v)
	t.signatureValues(signature)
	return t.serialize(chainId,true)
}

func (t *Transaction) serialize(chainId uint64, signature bool) (string, error) {

	// my brain has turned to mush, pretty sure this not the most efficient way
	// but trying to follow ether.js logic
	var list []rlp.Value
	list = append(list, t.Nonce.RLP())
	list = append(list, t.GasPrice.RLP())
	list = append(list, t.Gas.RLP())
	list = append(list, rlp.Value{String: t.To.String()})
	list = append(list, t.Value.RLP())
	//list = append(list, t.Input.Data.String())
	list = append(list, rlp.Value{String: t.Input.String()})

	if !signature {
		if chainId != 0 {
			list = append(list, QuantityFromUInt64(chainId).RLP())
			r, err := rlp.From("0x")
			if err != nil {
				return "", err
			}
			list = append(list, *r)
			s, err := rlp.From("0x")
			if err != nil {
				return "", err
			}
			list = append(list, *s)
		}
	} else {
		if chainId != 0 {
			v := chainId*2 + 8
			newV := t.V.UInt64() + v
			t.V = QuantityFromUInt64(newV)
		}
		log.Println("V: ", t.V.UInt64())
		log.Println("R: ", t.R.UInt64())
		log.Println("S: ", t.S.UInt64())
		list = append(list, t.V.RLP())
		list = append(list, t.R.RLP())
		list = append(list, t.S.RLP())
	}

	rawTx := rlp.Value{ List: list,}

	return rawTx.Encode()
}


func (t* Transaction) check() bool {
	// unsure if/how best to check other values
	if t.To == nil {
		return false
	}
	return true
}

// SignatureValues returns signature values. This signature
// needs to be in the [R || S || V] format where V is 0 or 1.
func (t* Transaction) signatureValues(sig *secp256k1.Signature) {

	t.R = QuantityFromBigInt(sig.R)
	t.S = QuantityFromBigInt(sig.S)
	t.V = QuantityFromUInt64(27)
}