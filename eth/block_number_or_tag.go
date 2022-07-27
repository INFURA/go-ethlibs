package eth

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type Tag string

const (
	// TagLatest aka head block
	TagLatest Tag = "latest"
	// TagEarliest aka genesis
	TagEarliest Tag = "earliest"
	// TagSafe lags behind "unsafe" by around 4 seconds and is less likely to reorg
	TagSafe Tag = "safe"
	// TagFinalized refers to a block that typically lags by one or two epochs (so 64-128 blocks)
	// but can lag further during consensus issues.  Once finalized a block can only reorg via a hard fork.
	TagFinalized Tag = "finalized"
	// TagPending refers to pending blocks. (Rarely used)
	TagPending Tag = "pending"
)

type BlockNumberOrTag struct {
	number Quantity
	tag    Tag
}

func (t Tag) String() string {
	return string(t)
}

func NewTag(s string) (*Tag, error) {
	switch s {
	case TagLatest.String(),
		TagEarliest.String(),
		TagPending.String(),
		TagSafe.String(),
		TagFinalized.String():
		t := Tag(s)
		return &t, nil
	default:
		return nil, errors.Errorf("invalid tag name %s", s)
	}
}

func MustTag(s string) *Tag {
	t, err := NewTag(s)
	if err != nil {
		panic(err)
	}
	return t
}

func NewBlockNumberOrTag(value string) (*BlockNumberOrTag, error) {
	/*
		The following options are possible for the defaultBlock parameter:

		HEX String - an integer block number
		String "earliest" for the earliest/genesis block
		String "latest" - for the latest mined block
		String "safe" - lags unsafe by around 4 seconds (less likely to reorg)
		String "finalized" -  lags by one or two epochs (so 64-128 blocks), ~~will never reorg~~.
		String "pending" - for the pending state/transactions
	*/

	b := BlockNumberOrTag{}

	switch value {
	case TagLatest.String(),
		TagEarliest.String(),
		TagPending.String(),
		TagSafe.String(),
		TagFinalized.String():
		b.tag = Tag(value)
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

func (b *BlockNumberOrTag) Tag() (Tag, bool) {
	if b == nil {
		return Tag(""), false
	}

	if b.tag == "" {
		return Tag(""), false
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

func (b BlockNumberOrTag) MarshalJSON() ([]byte, error) {
	if b.tag != "" {
		return json.Marshal(&b.tag)
	}

	return json.Marshal(&b.number)
}
