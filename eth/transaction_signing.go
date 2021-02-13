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

func (t *Transaction) Sign(privateKey string, chainId Quantity) (*Data, error) {

	if !t.check() {
		return nil, ErrInsufficientParams
	}

	pKey, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, err
	}

	// Get the data to sign, which is a hash of the type-dependent fields
	hash, err := t.SigningHash(chainId)

	// And sign the hash with the key
	signature, err := ECSign(hash, pKey, chainId)
	if err != nil {
		return nil, err
	}

	// Update signature values based on transaction type
	switch {
	case t.Type == nil:
		t.R, t.S, t.V = signature.EIP155Values()
	case t.Type.Int64() == TransactionTypeAccessList.Int64():
		// set RSV to EIP2718 values
		t.R, t.S, t.V = signature.EIP2718Values()
	default:
		return nil, errors.New("unsupported transaction type")
	}

	// And compute the raw representation of the tx
	raw, err := t.RawRepresentation(chainId)
	// TODO: update t.Raw ?
	return raw, err
}

func (t *Transaction) SigningHash(chainId Quantity) (*Hash, error) {
	switch {
	case t.Type == nil:
		// Legacy Transaction
		// Return Hash(RLP(Nonce, GasPrice, Gas, To, Value, Input, ChainId, 0, 0))
		zero := QuantityFromInt64(0)
		message := rlp.Value{List: []rlp.Value{
			t.Nonce.RLP(),
			t.GasPrice.RLP(),
			t.Gas.RLP(),
			t.To.RLP(),
			t.Value.RLP(),
			{String: t.Input.String()},
			chainId.RLP(),
			zero.RLP(),
			zero.RLP(),
		}}

		if s, err := message.Hash(); err != nil {
			return nil, err
		} else {
			return NewHash(s)
		}
	case t.Type.Int64() == TransactionTypeAccessList.Int64():
		// Return Hash( 0x1 || RLP(chainId, ...)
		payload := rlp.Value{List: []rlp.Value{
			chainId.RLP(),
			t.Nonce.RLP(),
			t.GasPrice.RLP(),
			t.Gas.RLP(),
			t.To.RLP(),
			t.Value.RLP(),
			{String: t.Input.String()},
			t.AccessList.RLP(),
		}}
		encoded, err := payload.Encode()
		if err != nil {
			return nil, err
		}
		data, err := NewData("0x01" + encoded[2:])
		if err != nil {
			return nil, err
		}
		h := data.Hash()
		return &h, nil
	}

	return nil, errors.New("unsupported transaction type")
}

func (t *Transaction) RawRepresentation(chainId Quantity) (*Data, error) {
	switch {
	case t.Type == nil:
		// Legacy Transactions are RLP(Nonce, GasPrice, Gas, To, Value, Input, V, R, S)
		message := rlp.Value{List: []rlp.Value{
			t.Nonce.RLP(),
			t.GasPrice.RLP(),
			t.Gas.RLP(),
			t.To.RLP(),
			t.Value.RLP(),
			{String: t.Input.String()},
			t.V.RLP(),
			t.R.RLP(),
			t.S.RLP(),
		}}
		if encoded, err := message.Encode(); err != nil {
			return nil, err
		} else {
			return NewData(encoded)
		}
	case t.Type.Int64() == TransactionTypeAccessList.Int64():
		// EIP-2930 Transactions are 0x1 || rlp([chainId, nonce, gasPrice, gasLimit, to, value, data, access_list, yParity, senderR, senderS])
		typePrefix, err := t.Type.RLP().Encode()
		if err != nil {
			return nil, err
		}
		payload := rlp.Value{List: []rlp.Value{
			chainId.RLP(),
			t.Nonce.RLP(),
			t.GasPrice.RLP(),
			t.Gas.RLP(),
			t.To.RLP(),
			t.Value.RLP(),
			{String: t.Input.String()},
			t.AccessList.RLP(),
			t.V.RLP(),
			t.R.RLP(),
			t.S.RLP(),
		}}
		if encodedPayload, err := payload.Encode(); err != nil {
			return nil, err
		} else {
			return NewData(typePrefix + encodedPayload[2:])
		}
	}

	return nil, errors.New("unsupported transaction type")
}

func (a AccessList) RLP() rlp.Value {
	val := rlp.Value{List: make([]rlp.Value, len(a))}
	for i := range a {
		keys := rlp.Value{List: make([]rlp.Value, len(a[i].StorageKeys))}
		for j, k := range a[i].StorageKeys {
			keys.List[j] = k.RLP()
		}
		val.List[i] = rlp.Value{List: []rlp.Value{
			a[i].Address.RLP(),
			keys,
		}}
	}
	return val
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

func (t *Transaction) serializeRaw(chainID uint64) (rlp.Value, error) {
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
	if t.To == nil {
		return false
	}
	return true
}
