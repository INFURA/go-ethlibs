package eth

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/INFURA/go-ethlibs/rlp"
)

type Condition json.RawMessage
type AccessList []AccessListEntry

var (
	TransactionTypeLegacy     = int64(0x0)
	TransactionTypeAccessList = int64(0x1)
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

type AccessListEntry struct {
	Address     Address  `json:"address"`
	StorageKeys []Data32 `json:"storageKeys"`
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
	if t.transactionType() == TransactionTypeLegacy {
		t.AccessList = nil
	}

	return nil
}

func (t *Transaction) transactionType() int64 {
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

func NewAccessListFromRLP(v rlp.Value) (AccessList, error) {
	accessList := make(AccessList, len(v.List))
	for j, accessRLP := range v.List {
		l := len(accessRLP.List)
		if l == 0 || l > 2 {
			return nil, errors.Errorf("invalid access list entry %d", j)
		}
		address, err := NewAddress(accessRLP.List[0].String)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid access list entry address %d", j)
		}
		accessList[j].Address = *address
		if l == 2 {
			// 2nd item is the storage keys
			accessList[j].StorageKeys = make([]Data32, len(accessRLP.List[1].List))
			for k, key := range accessRLP.List[1].List {
				d, err := NewData32(key.String)
				if err != nil {
					return nil, errors.Wrapf(err, "invalid access list entry %d storage key %d", j, k)
				}
				accessList[j].StorageKeys[k] = *d
			}
		}
	}

	return accessList, nil
}
