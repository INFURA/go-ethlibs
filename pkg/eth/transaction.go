package eth

import (
	"encoding/json"
)

type Condition json.RawMessage

type Transaction struct {
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
	Creates   *Hash      `json:"creates,omitempty"`
	Condition *Condition `json:"condition,omitempty"`

	// Keep the source so we can recreate its expected representation
	source string `json:"-"`
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

	return nil
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	if t.source == "parity" {
		type parity struct {
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
			Creates   *Hash      `json:"creates"`
			Condition *Condition `json:"condition"`
		}

		p := parity{
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
