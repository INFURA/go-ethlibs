package eth

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/INFURA/go-ethlibs/rlp"
)

var ErrInsufficientParams = errors.New("transaction is missing values")

// Looks like there have been multiple attempts to get the Koblitz curve (secp256k1) supported in golang
//
// https://github.com/golang/go/pull/26873 <-- rejected
// https://github.com/golang/go/issues/26776 <-- rejected
// using "github.com/btcsuite/btcd/btcec"
// other alternative is to import the C library but want to avoid that if possible

// Sign uses the hex-encoded private key and chainId to update the R, S, and V values
// for a Transaction, and returns the raw signed transaction or an error.
func (t *Transaction) Sign(privateKey string, chainId Quantity) (*Data, error) {
	var (
		pKey []byte
		err  error
	)

	if strings.HasPrefix(privateKey, "0x") && len(privateKey) > 2 {
		pKey, err = hex.DecodeString(privateKey[2:])
	} else {
		pKey, err = hex.DecodeString(privateKey)
	}

	if err != nil {
		return nil, err
	}

	// Get the data to sign, which is a hash of the type-dependent fields
	hash, err := t.SigningHash(chainId)
	if err != nil {
		return nil, err
	}

	// And sign the hash with the key
	signature, err := ECSign(hash, pKey, chainId)
	if err != nil {
		return nil, err
	}

	// Update signature values based on transaction type
	switch t.TransactionType() {
	case TransactionTypeLegacy:
		t.R, t.S, t.V = signature.EIP155Values()
	case TransactionTypeAccessList:
		// set chainId and RSV to EIP2718 values
		t.ChainId = &chainId
		t.R, t.S, t.V = signature.EIP2718Values()
	default:
		return nil, errors.New("unsupported transaction type")
	}

	// And compute the raw representation of the tx
	raw, err := t.RawRepresentation()
	if err != nil {
		return nil, err
	}
	if t.Raw != nil {
		// Update .Raw to ensure it matches (currently only provided for Parity-flavored txs)
		t.Raw = raw
	}

	// recover the sender as well
	signingHash, err := t.SigningHash(chainId)
	if err != nil {
		return nil, err
	}

	sender, err := signature.Recover(signingHash)
	if err != nil {
		return nil, err
	}

	t.From = *sender
	t.Hash = raw.Hash()
	return raw, err
}

// SigningPreimage returns the opaque data preimage that is required for signing a given transaction type
func (t *Transaction) SigningPreimage(chainId Quantity) (*Data, error) {
	switch t.TransactionType() {
	case TransactionTypeLegacy:
		var message rlp.Value

		// If chainId is zero then this transaction is NOT EIP-155 protected
		if chainId.Int64() == 0 {
			// In that case we should sign
			// RLP(Nonce, GasPrice, Gas, To, Value, Input)
			message = rlp.Value{List: []rlp.Value{
				t.Nonce.RLP(),
				t.GasPrice.RLP(),
				t.Gas.RLP(),
				t.To.RLP(),
				t.Value.RLP(),
				{String: t.Input.String()},
			}}
		} else {
			// On the other hand, for EIP-155 txs we should sign
			// RLP(Nonce, GasPrice, Gas, To, Value, Input, ChainId, 0, 0)
			zero := QuantityFromInt64(0)
			message = rlp.Value{List: []rlp.Value{
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
		}

		// encode the list as RLP
		encoded, err := message.Encode()
		if err != nil {
			return nil, err
		}
		// and return it
		return NewData(encoded)
	case TransactionTypeAccessList:
		// Return 0x1 || RLP(chainId, Nonce, GasPrice, Gas, To, Value, Input, AccessList)
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
		// encode the list as RLP
		encoded, err := payload.Encode()
		if err != nil {
			return nil, err
		}
		// And return it with the 0x01 prefix
		return NewData("0x01" + encoded[2:])
	default:
		return nil, errors.New("unsupported transaction type")
	}
}

// SigningHash returns the Keccak-256 hash of the transaction fields required for transaction signing or an error.
func (t *Transaction) SigningHash(chainId Quantity) (*Hash, error) {
	// Get the opaque preimage
	preimage, err := t.SigningPreimage(chainId)
	if err != nil {
		return nil, err
	}

	// And return the preimage's hash
	h := preimage.Hash()
	return &h, nil
}

// RawRepresentation returns the transaction encoded as a raw hexadecimal data string, or an error
func (t *Transaction) RawRepresentation() (*Data, error) {
	switch t.TransactionType() {
	case TransactionTypeLegacy:
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
	case TransactionTypeAccessList:
		// EIP-2930 Transactions are 0x1 || rlp([chainId, nonce, gasPrice, gasLimit, to, value, data, access_list, yParity, senderR, senderS])
		typePrefix, err := t.Type.RLP().Encode()
		if err != nil {
			return nil, err
		}
		if t.ChainId == nil {
			return nil, errors.New("chainID is required on EIP-2930 transactions")
		}
		payload := rlp.Value{List: []rlp.Value{
			t.ChainId.RLP(),
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
	default:
		return nil, errors.New("unsupported transaction type")
	}
}

// Signature returns the R, S, V values for a transaction, and the ChainId if available
func (t *Transaction) Signature() (*Signature, error) {
	switch t.TransactionType() {
	case TransactionTypeLegacy:
		return NewEIP155Signature(t.R, t.S, t.V)
	case TransactionTypeAccessList:
		if t.ChainId == nil {
			return nil, errors.New("chainId is required")
		}
		return NewEIP2718Signature(*t.ChainId, t.R, t.S, t.V)
	default:
		return nil, errors.New("unsupported transaction type")
	}
}

// IsProtected returns true if a transaction is replay protected, either via EIP-155 or newer transaction formats.
// This method returns false for transactions with invalid signatures.
func (t *Transaction) IsProtected() bool {
	signature, err := t.Signature()
	if err != nil {
		return false
	}
	return signature.chainId.Int64() != 0x0
}
