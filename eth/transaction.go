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
)

type Transaction struct {
	Type        *Quantity `json:"type,omitempty"`
	BlockHash   *Hash     `json:"blockHash"`
	BlockNumber *Quantity `json:"blockNumber"`
	From        Address   `json:"from"`
	Gas         Quantity  `json:"gas"`
	Hash        Hash      `json:"hash"`
	Input       Data      `json:"input"`
	Nonce       Quantity  `json:"nonce"`
	To          *Address  `json:"to"`
	Index       *Quantity `json:"transactionIndex"`
	Value       Quantity  `json:"value"`
	V           Quantity  `json:"v"`
	R           Quantity  `json:"r"`
	S           Quantity  `json:"s"`
	Data        []byte    `json:"data"`

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

	// Keep the source so we can recreate its expected representation
	source string
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
		return nil
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
	}

	if len(fields) > 0 {
		return fmt.Errorf("missing required field(s) %s for transaction type", strings.Join(fields, ","))
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
	case TransactionTypeDynamicFee:
		// We introduce a new EIP-2718 transaction type, with the format 0x02 || rlp([chainId, nonce, maxPriorityFeePerGas, maxFeePerGas, gasLimit, to, value, data, access_list, signatureYParity, signatureR, signatureS]).
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

func (t *Transaction) MarshalJSON() ([]byte, error) {
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
			Input       Data      `json:"input"`
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
		g := geth(*t)
		return json.Marshal(&g)
	}

	type unknown Transaction
	u := unknown(*t)
	return json.Marshal(&u)
}
