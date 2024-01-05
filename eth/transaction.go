package eth

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/INFURA/go-ethlibs/rlp"
)

type Condition json.RawMessage

var (
	TransactionTypeLegacy     = int64(0x0) // TransactionTypeLegacy refers to pre-EIP-2718 transactions.
	TransactionTypeAccessList = int64(0x1) // TransactionTypeAccessList refers to EIP-2930 transactions.
	TransactionTypeDynamicFee = int64(0x2) // TransactionTypeDynamicFee refers to EIP-1559 transactions.
	TransactionTypeBlob       = int64(0x3) // TransactionTypeBlob refers to EIP-4844 "blob" transactions.
)

type Transaction struct {
	Type        *Quantity `json:"type,omitempty"`
	BlockHash   *Hash     `json:"blockHash"`
	BlockNumber *Quantity `json:"blockNumber"`
	From        Address   `json:"from"`
	Gas         Quantity  `json:"gas"`
	Hash        Hash      `json:"hash"`
	Input       Input     `json:"input"`
	Nonce       Quantity  `json:"nonce"`
	To          *Address  `json:"to"`
	Index       *Quantity `json:"transactionIndex"`
	Value       Quantity  `json:"value"`
	V           Quantity  `json:"v"`
	R           Quantity  `json:"r"`
	S           Quantity  `json:"s"`
	YParity     *Quantity `json:"yParity,omitempty"`

	// Gas Price (optional since not included in EIP-1559)
	GasPrice *Quantity `json:"gasPrice,omitempty"`

	// EIP-1559 MaxFeePerGas/MaxPriorityFeePerGas (optional since only included in EIP-1559 transactions)
	MaxFeePerGas         *Quantity `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas *Quantity `json:"maxPriorityFeePerGas,omitempty"`

	// Parity Fields
	StandardV *Quantity  `json:"standardV,omitempty"`
	Raw       *Data      `json:"raw,omitempty"`
	PublicKey *Data      `json:"publicKey,omitempty"`
	ChainId   *Quantity  `json:"chainId,omitempty"`
	Creates   *Address   `json:"creates,omitempty"` // Parity wiki claims this is a Hash
	Condition *Condition `json:"condition,omitempty"`

	// EIP-2930 accessList
	AccessList *AccessList `json:"accessList,omitempty"`

	// EIP-4844 blob fields
	MaxFeePerBlobGas    *Quantity `json:"maxFeePerBlobGas,omitempty"`
	BlobVersionedHashes Hashes    `json:"blobVersionedHashes,omitempty"`

	// EIP-4844 Blob transactions in "Network Representation" include the additional
	// fields from the BlobsBundleV1 engine API schema.  However, these fields are not
	// available at the execution layer and thus not expected to be seen when
	// dealing with JSONRPC representations of transactions, and are excluded from
	// JSON Marshalling.  As such, this field is only populated when decoding a
	// raw transaction in "Network Representation" and the fields must be accessed directly.
	BlobBundle *BlobsBundleV1 `json:"-"`

	// Keep the source so we can recreate its expected representation
	source string
}

type BlobsBundleV1 struct {
	Blobs       []Data `json:"blobs"`
	Commitments []Data `json:"commitments"`
	Proofs      []Data `json:"proofs"`
}

type NewPendingTxBodyNotificationParams struct {
	Subscription string      `json:"subscription"`
	Result       Transaction `json:"result"`
}

type NewPendingTxNotificationParams struct {
	Subscription string `json:"subscription"`
	Result       Hash   `json:"result"`
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	type tx Transaction
	aliased := tx(*t)

	err := json.Unmarshal(data, &aliased)
	if err != nil {
		return err
	}

	*t = Transaction(aliased)
	if t.StandardV != nil {
		t.source = "parity"
	} else {
		t.source = "geth"
	}

	// Force AccessList to nil for legacy txs
	if t.TransactionType() == TransactionTypeLegacy {
		t.AccessList = nil
	}

	return nil
}

// TransactionType returns the transactions EIP-2718 type, or TransactionTypeLegacy for pre-EIP-2718 transactions.
func (t *Transaction) TransactionType() int64 {
	if t.Type == nil {
		return TransactionTypeLegacy
	}

	return t.Type.Int64()
}

// RequiredFields inspects the Transaction Type and returns an error if any required fields are missing
func (t *Transaction) RequiredFields() error {
	var fields []string
	switch t.TransactionType() {
	case TransactionTypeLegacy:
		if t.GasPrice == nil {
			fields = append(fields, "gasPrice")
		}
	case TransactionTypeAccessList:
		if t.ChainId == nil {
			fields = append(fields, "chainId")
		}
	case TransactionTypeDynamicFee:
		if t.ChainId == nil {
			fields = append(fields, "chainId")
		}
		if t.MaxFeePerGas == nil {
			fields = append(fields, "maxFeePerGas")
		}
		if t.MaxPriorityFeePerGas == nil {
			fields = append(fields, "maxPriorityFeePerGas")
		}
	case TransactionTypeBlob:
		if t.ChainId == nil {
			fields = append(fields, "chainId")
		}
		if t.MaxFeePerBlobGas == nil {
			fields = append(fields, "maxFeePerBlobGas")
		}
		if t.BlobVersionedHashes == nil {
			fields = append(fields, "blobVersionedHashes")
		}
		if t.To == nil {
			// Contract creation not supported in blob txs
			fields = append(fields, "to")
		}
	default:
		return errors.New("unsupported transaction type")
	}

	if len(fields) > 0 {
		return fmt.Errorf("missing required field(s) %s for transaction type %d", strings.Join(fields, ","), t.TransactionType())
	}

	return nil
}

// RawRepresentation returns the transaction encoded as a raw hexadecimal data string, or an error
func (t *Transaction) RawRepresentation() (*Data, error) {
	if err := t.RequiredFields(); err != nil {
		return nil, err
	}

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
		var yParity Quantity
		if t.YParity != nil {
			yParity = *t.YParity
		} else {
			yParity = t.V
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
			yParity.RLP(),
			t.R.RLP(),
			t.S.RLP(),
		}}
		if encodedPayload, err := payload.Encode(); err != nil {
			return nil, err
		} else {
			return NewData(typePrefix + encodedPayload[2:])
		}
	case TransactionTypeDynamicFee:
		// We introduce a new EIP-2718 transaction type, with the format 0x02 || rlp([chainId, nonce, maxPriorityFeePerGas, maxFeePerGas, gasLimit, to, value, data, access_list, signatureYParity, signatureR, signatureS]).
		typePrefix, err := t.Type.RLP().Encode()
		if err != nil {
			return nil, err
		}
		var yParity Quantity
		if t.YParity != nil {
			yParity = *t.YParity
		} else {
			yParity = t.V
		}
		payload := rlp.Value{List: []rlp.Value{
			t.ChainId.RLP(),
			t.Nonce.RLP(),
			t.MaxPriorityFeePerGas.RLP(),
			t.MaxFeePerGas.RLP(),
			t.Gas.RLP(),
			t.To.RLP(),
			t.Value.RLP(),
			{String: t.Input.String()},
			t.AccessList.RLP(),
			yParity.RLP(),
			t.R.RLP(),
			t.S.RLP(),
		}}
		if encodedPayload, err := payload.Encode(); err != nil {
			return nil, err
		} else {
			return NewData(typePrefix + encodedPayload[2:])
		}
	case TransactionTypeBlob:
		// We introduce a new EIP-2718 transaction, “blob transaction”, where the TransactionType is BLOB_TX_TYPE and the TransactionPayload is the RLP serialization of the following TransactionPayloadBody:
		//[chain_id, nonce, max_priority_fee_per_gas, max_fee_per_gas, gas_limit, to, value, data, access_list, max_fee_per_blob_gas, blob_versioned_hashes, y_parity, r, s]
		typePrefix, err := t.Type.RLP().Encode()
		if err != nil {
			return nil, err
		}
		payload := rlp.Value{List: []rlp.Value{
			t.ChainId.RLP(),
			t.Nonce.RLP(),
			t.MaxPriorityFeePerGas.RLP(),
			t.MaxFeePerGas.RLP(),
			t.Gas.RLP(),
			t.To.RLP(),
			t.Value.RLP(),
			{String: t.Input.String()},
			t.AccessList.RLP(),
			t.MaxFeePerBlobGas.RLP(),
			t.BlobVersionedHashes.RLP(),
			t.YParity.RLP(),
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

// NetworkRepresentation returns the transaction encoded as a raw hexadecimal data string suitable
// for network transmission, or an error.  For Blob transactions this includes the blob payload.
func (t *Transaction) NetworkRepresentation() (*Data, error) {
	if err := t.RequiredFields(); err != nil {
		return nil, err
	}

	switch t.TransactionType() {
	case TransactionTypeLegacy, TransactionTypeAccessList, TransactionTypeDynamicFee:
		// For most transaction types, the "Raw" and "Network" representations are the same
		return t.RawRepresentation()
	case TransactionTypeBlob:
		// Blob transactions have two network representations. During transaction gossip responses (PooledTransactions),
		// the EIP-2718 TransactionPayload of the blob transaction is wrapped to become:
		//
		// rlp([tx_payload_body, blobs, commitments, proofs])
		typePrefix, err := t.Type.RLP().Encode()
		if err != nil {
			return nil, err
		}

		if t.BlobBundle == nil {
			return nil, errors.New("network representation of blob txs requires populated blob data")
		}

		body := rlp.Value{List: []rlp.Value{
			t.ChainId.RLP(),
			t.Nonce.RLP(),
			t.MaxPriorityFeePerGas.RLP(),
			t.MaxFeePerGas.RLP(),
			t.Gas.RLP(),
			t.To.RLP(),
			t.Value.RLP(),
			{String: t.Input.String()},
			t.AccessList.RLP(),
			t.MaxFeePerBlobGas.RLP(),
			t.BlobVersionedHashes.RLP(),
			t.YParity.RLP(),
			t.R.RLP(),
			t.S.RLP(),
		}}
		dataList := func(data []Data) rlp.Value {
			v := rlp.Value{
				List: make([]rlp.Value, 0, len(data)),
			}
			for i := range data {
				v.List = append(v.List, data[i].RLP())
			}
			return v
		}
		blobs := dataList(t.BlobBundle.Blobs)
		commitments := dataList(t.BlobBundle.Commitments)
		proofs := dataList(t.BlobBundle.Proofs)

		payload := rlp.Value{List: []rlp.Value{
			body,
			blobs,
			commitments,
			proofs,
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

func (t Transaction) MarshalJSON() ([]byte, error) {
	if t.source == "parity" {
		type parity struct {
			// TODO: Revisit once open ethereuem w/ EIP-2930 is released
			Type        *Quantity `json:"type,omitempty"` // FIXME: current OE uses `int` instead of QUANTITY encoding
			BlockHash   *Hash     `json:"blockHash"`
			BlockNumber *Quantity `json:"blockNumber"`
			From        Address   `json:"from"`
			Gas         Quantity  `json:"gas"`
			GasPrice    *Quantity `json:"gasPrice"`
			Hash        Hash      `json:"hash"`
			Input       Input     `json:"input"`
			Nonce       Quantity  `json:"nonce"`
			To          *Address  `json:"to"`
			Index       *Quantity `json:"transactionIndex"`
			Value       Quantity  `json:"value"`
			V           Quantity  `json:"v"`
			R           Quantity  `json:"r"`
			S           Quantity  `json:"s"`

			// Parity Fields
			StandardV *Quantity  `json:"standardV"`
			Raw       *Data      `json:"raw"`
			PublicKey *Data      `json:"publicKey"`
			ChainId   *Quantity  `json:"chainId"`
			Creates   *Address   `json:"creates"`
			Condition *Condition `json:"condition"`

			AccessList *AccessList `json:"accessList,omitempty"`
		}

		p := parity{
			Type:        t.Type,
			BlockHash:   t.BlockHash,
			BlockNumber: t.BlockNumber,
			From:        t.From,
			Gas:         t.Gas,
			GasPrice:    t.GasPrice,
			Hash:        t.Hash,
			Input:       t.Input,
			Nonce:       t.Nonce,
			To:          t.To,
			Index:       t.Index,
			Value:       t.Value,
			V:           t.V,
			R:           t.R,
			S:           t.S,

			// Parity Fields
			StandardV: t.StandardV,
			Raw:       t.Raw,
			PublicKey: t.PublicKey,
			ChainId:   t.ChainId,
			Creates:   t.Creates,
			Condition: t.Condition,

			AccessList: t.AccessList,
		}

		return json.Marshal(&p)

	} else if t.source == "geth" {
		type geth Transaction
		g := geth(t)
		return json.Marshal(&g)
	}

	type unknown Transaction
	u := unknown(t)
	return json.Marshal(&u)
}
