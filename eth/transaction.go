package eth

import (
	"encoding/json"
)

type Condition json.RawMessage

var (
	TransactionTypeLegacy     = int64(0x0) // TransactionTypeLegacy refers to pre-EIP-2718 transactions.
	TransactionTypeAccessList = int64(0x1) // TransactionTypeAccessList refers to EIP-2930 transactions.
)

type Transaction struct {
	Type        *Quantity `json:"type,omitempty"`
	BlockHash   *Hash     `json:"blockHash"`
	BlockNumber *Quantity `json:"blockNumber"`
	From        Address   `json:"from"`
	Gas         Quantity  `json:"gas"`
	GasPrice    Quantity  `json:"gasPrice"`
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

func (t *Transaction) MarshalJSON() ([]byte, error) {
	if t.source == "parity" {
		type parity struct {
			// TODO: Revisit once open ethereuem w/ EIP-2930 is released
			Type        *Quantity `json:"type,omitempty"` // FIXME: current OE uses `int` instead of QUANTITY encoding
			BlockHash   *Hash     `json:"blockHash"`
			BlockNumber *Quantity `json:"blockNumber"`
			From        Address   `json:"from"`
			Gas         Quantity  `json:"gas"`
			GasPrice    Quantity  `json:"gasPrice"`
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
