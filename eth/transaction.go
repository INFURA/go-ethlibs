package eth

import (
	"encoding/json"
	"errors"
)

type Condition json.RawMessage


// require values for valid tx
func NewTransaction(nonce int64, gasPrice int64, gasLimit int64, toAddr string, value int64,  data []byte/*, chainId int64*/) (*Transaction, error) {
	//chainID := QuantityFromInt64(chainId)
	t := Transaction{
		Nonce: QuantityFromInt64(nonce),
		GasPrice: QuantityFromInt64(gasPrice),
		Gas: QuantityFromInt64(gasLimit),
		To: MustAddress(toAddr),
		Value: QuantityFromInt64(value),
		Input: *MustData(string(data)),
		//ChainId: &chainID,
	}
	return &t, nil
}

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
	Creates   *Address   `json:"creates,omitempty"` // Parity wiki claims this is a Hash
	Condition *Condition `json:"condition,omitempty"`

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

var ErrInsufficientParams = errors.New("transaction is missing values")

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
			Creates   *Address   `json:"creates"`
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
