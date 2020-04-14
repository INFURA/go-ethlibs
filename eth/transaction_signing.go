package eth

import (
	"encoding/hex"
	"errors"

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

	rawTx, err := t.serialize(chainId, false)
	if err != nil {
		return "", err
	}

	privKey, pubKey := secp256k1.PrivKeyFromBytes(secp256k1.S256(), pKey)
	signature, err := privKey.Sign(rawTx)
	if err != nil {
		return "", err
	}

	verified := signature.Verify(rawTx, pubKey)
	if !verified {
		panic(ErrFatalCrytpo)
	}

	// extract signature (r,s,v)
	t.signatureValues(signature)
	signed, err := t.serialize(chainId, true)
	return string(signed), err
}

func (t *Transaction) serialize(chainId uint64, signature bool) ([]byte, error) {

	var list []rlp.Value
	list = append(list, t.Nonce.RLP())
	list = append(list, t.GasPrice.RLP())
	list = append(list, t.Gas.RLP())
	list = append(list, t.To.RLP())
	list = append(list, t.Value.RLP())
	list = append(list, rlp.Value{String: t.Input.String()})

	var temp []byte
	if !signature {
		if chainId != 0 {
			list = append(list, QuantityFromUInt64(chainId).RLP())
			r, err := rlp.From("0x")
			if err != nil {
				return temp, err
			}
			list = append(list, *r)
			s, err := rlp.From("0x")
			if err != nil {
				return temp, err
			}
			list = append(list, *s)
		}
		rawTx := rlp.Value{List: list}
		h, err := rawTx.HashToBytes()
		return h, err
	} else {
		if chainId != 0 {
			v := chainId*2 + 8
			newV := t.V.UInt64() + v
			t.V = QuantityFromUInt64(newV)
		}
		list = append(list, t.V.RLP())
		list = append(list, t.R.RLP())
		list = append(list, t.S.RLP())
		rawTx := rlp.Value{List: list}

		signed, err := rawTx.Encode()
		return []byte(signed), err
	}
}

func (t *Transaction) check() bool {
	// unsure if/how best to check other values
	if t.To == nil {
		return false
	}
	return true
}

// EIP 155 (https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md)
// says using 27 or 28 (or 35 or 36 depending on how you look at it) for V value are correct
// I am using 27 but the python library I've been using is using 28.
// It looks like ether.js
// https://github.com/ethers-io/ethers.js/blob/b1c6575a1b8cc666a9173eceedb7a367329819c7/dist/ethers.js#L14074
// is using the 28 value as some kind of recoveryParam but unsure as the extra v seems to be filled out by the signer
// https://github.com/ethereumjs/ethereumjs-tx/blob/b564c15e3eb709d1a677cac25c88d670b5ff0e01/src/transaction.ts#L262
// using 27 as our signer doesn't have any V value it can return
func (t *Transaction) signatureValues(sig *secp256k1.Signature) {
	t.R = QuantityFromBigInt(sig.R)
	t.S = QuantityFromBigInt(sig.S)
	t.V = QuantityFromUInt64(28)
}
