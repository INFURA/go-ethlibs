package eth

import (
	"fmt"
	secp256k1 "github.com/btcsuite/btcd/btcec"
	"github.com/INFURA/go-ethlibs/rlp"
	"log"
	"math/big"
)


// Looks like there have been multiple attempts to get the Koblitz curve (secp256k1) supported in golang
//
// https://github.com/golang/go/pull/26873 <-- rejected
// https://github.com/golang/go/issues/26776 <-- rejected
// using "github.com/btcsuite/btcd/btcec"
// other alternative is to import the C library but want to avoid that if possible


func (t *Transaction) Sign(privateKey string, chainId uint64) (string, error) {

	// add ChainID for easy signing..  not sure this is best but makes the serializing/signing easier
	c := QuantityFromUInt64(chainId)
	t.ChainId = &c
	if !t.check() {
		return "", ErrInsufficientParams
	}

	rawTx, err := t.serialize(false)
	if err != nil {
		return "", err
	}

	privKey, pubKey := secp256k1.PrivKeyFromBytes(secp256k1.S256(), []byte(privateKey))
	signature, err := privKey.Sign([]byte(rawTx))
	if err != nil {
		return "", err
	}
	//return t.serializeWithSignature(signature)
	verified := signature.Verify([]byte(rawTx), pubKey)
	log.Println(verified)

	// extract signature (r,s,v)
	r,s,v,err := signatureValues(signature.Serialize())
	if err != nil {
		return "", err
	}
	t.R = QuantityFromBigInt(r)
	t.S = QuantityFromBigInt(s)
	t.V = QuantityFromBigInt(v)

	return t.serialize(true)
	//return signature, nil
}

func (t *Transaction) serialize(signature bool) (string, error) {

	// my brain has turned to mush, pretty sure this not the most efficient way
	// but trying to follow ether.js logic
	var list []rlp.Value
	list = append(list, t.Nonce.RLP())
	list = append(list, t.GasPrice.RLP())
	list = append(list, t.Gas.RLP())
	list = append(list, t.To.RLP())
	list = append(list, t.Value.RLP())
	//list = append(list, t.Input.Data.String())
	list = append(list, rlp.Value{String: t.Input.String()})

	if !signature {
		/*
		if t.ChainId != nil && t.ChainId.UInt64() != 0 {
			list = append(list, t.ChainId.RLP())
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

		 */
	} else {
		if t.ChainId != nil && t.ChainId.UInt64() != 0 {
			v := t.ChainId.UInt64()*2 + 8
			newV := t.V.UInt64() + v
			t.V = QuantityFromUInt64(newV)
		}
		list = append(list, t.V.RLP())
		list = append(list, t.R.RLP())
		list = append(list, t.S.RLP())
	}

	rawTx := rlp.Value{ List: list,}
	for _, val := range list {
		log.Println(val.String)
	}
	return rawTx.Encode()
}


func (t* Transaction) check() bool {
	// unsure if/how best to check other values
	if t.To == nil || t.ChainId == nil {
		return false
	}
	return true
}

// SignatureValues returns signature values. This signature
// needs to be in the [R || S || V] format where V is 0 or 1.
func signatureValues(sig []byte) (r, s, v *big.Int, err error) {
	if len(sig) != 65 {
		panic(fmt.Sprintf("wrong size for signature: got %d, want 65", len(sig)))
	}
	r = new(big.Int).SetBytes(sig[:32])
	s = new(big.Int).SetBytes(sig[32:64])
	v = new(big.Int).SetBytes([]byte{sig[64] + 27})

	return r, s, v, nil
}