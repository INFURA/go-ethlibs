package eth

import (
	"encoding/json"
)

const (
	TagLatest   = "latest"
	TagEarliest = "earliest"
	TagPending  = "pending"
)

type BlockNumberOrTag struct {
	number Quantity
	tag    string
}

func NewBlockNumberOrTag(value string) (*BlockNumberOrTag, error) {
	/*
		The following options are possible for the defaultBlock parameter:

		HEX String - an integer block number
		String "earliest" for the earliest/genesis block
		String "latest" - for the latest mined block
		String "pending" - for the pending state/transactions
	*/

	b := BlockNumberOrTag{}

	switch value {
	case TagLatest, TagEarliest, TagPending:
		b.tag = value
		return &b, nil
	default:
		q, err := NewQuantity(value)
		if err != nil {
			return nil, err
		}
		b.number = *q
		return &b, nil
	}
}

func MustBlockNumberOrTag(value string) *BlockNumberOrTag {
	b, err := NewBlockNumberOrTag(value)
	if err != nil {
		panic(err)
	}

	return b
}

func (b *BlockNumberOrTag) Tag() (string, bool) {
	if b == nil {
		return "", false
	}

	if b.tag == "" {
		return "", false
	}

	return b.tag, true
}

func (b *BlockNumberOrTag) Quantity() (Quantity, bool) {
	if b == nil {
		return Quantity{}, false
	}

	if b.tag == "" {
		return b.number, true
	}

	return Quantity{}, false
}

func (b *BlockNumberOrTag) UnmarshalJSON(data []byte) error {
	str := ""
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}

	_b, err := NewBlockNumberOrTag(str)
	if err != nil {
		return err
	}

	*b = *_b
	return nil
}

func (b *BlockNumberOrTag) MarshalJSON() ([]byte, error) {
	if b.tag != "" {
		return json.Marshal(&b.tag)
	}

	return json.Marshal(&b.number)
}
