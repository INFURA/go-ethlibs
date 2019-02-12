package eth

import (
	"encoding/json"
)

type Condition json.RawMessage

type Transaction struct {
	BlockHash   Hash        `json:"blockHash"`
	BlockNumber BlockNumber `json:"blockNumber"`
	From        Address     `json:"from"`
	Gas         Quantity    `json:"gas"`
	GasPrice    Quantity    `json:"gasPrice"`
	Hash        Hash        `json:"Hash"`
	Input       Data        `json:"input"`
	Nonce       Quantity    `json:"nonce"`
	To          Address     `json:"To"`
	Index       *Quantity   `json:"transactionIndex"`
	Value       Quantity    `json:"value"`
	V           Quantity    `json:"v"`
	R           Data32      `json:"r"`
	S           Data32      `json:"s"`

	// Parity Fields
	StandardV Quantity  `json:"standardV,omitempty"`
	Raw       Data      `json:"raw,omitempty"`
	PublicKey Hash      `json:"publicKey,omitempty"`
	ChainId   Quantity  `json:"chainId,omitempty"`
	Creates   Hash      `json:"creates,omitempty"`
	Condition Condition `json:"condition,omitempty"`
}
