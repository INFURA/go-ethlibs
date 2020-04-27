package eth

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type BlockSpecifier struct {
	Number           *Quantity
	Tag              *Tag
	Hash             *Hash
	RequireCanonical *bool
	Raw              bool
}

func NewBlockSpecifierFromString(value string) (*BlockSpecifier, error) {
	dec, err := NewBlockNumberOrTag(value)
	if err != nil {
		return nil, err
	}
	b := BlockSpecifier{}
	if tag, found := dec.Tag(); found {
		b.Tag = &tag
	}
	if num, found := dec.Quantity(); found {
		b.Number = &num
	}
	return &b, nil
}

func NewBlockSpecifierFromMap(value map[string]interface{}) (*BlockSpecifier, error) {
	b := BlockSpecifier{}
	if h, found := value["blockHash"]; found {
		// "blockHash" takes precendence over "blockNumber"
		hash, err := NewHash(h.(string))
		if err != nil {
			return nil, err
		}
		b.Hash = hash

		// set the "requireCanonical" flag (default false)
		canonical := false
		if cn, found := value["requireCanonical"]; found {
			switch cn.(type) {
			case bool:
				canonical = cn.(bool)
			default:
				return nil, errors.New(`"requireCanonical" must be a boolean value`)
			}
		}
		b.RequireCanonical = &canonical

	} else if n, found := value["blockNumber"]; found {
		num, err := NewQuantity(n.(string))
		if err != nil {
			return nil, err
		}
		b.Number = num

	} else {
		return nil, errors.New(`expected either a "blockHash" or a "blockNumber" value`)
	}
	return &b, nil
}

func NewBlockSpecifier(value interface{}) (*BlockSpecifier, error) {
	/*
		The following values are possible according to EIP-1898:
		- a block tag (string): "earliest", "latest", "pending"
		- a block number (hex string): "0x0"
		- a block specifier (map):
			{ "blockNumber": "0x0" }
			{ "blockHash": "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3" }
			{ "blockHash": "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3", "requireCanonical": true }
			{ "blockHash": "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3", "requireCanonical": false }
	*/
	switch value.(type) {
	case string:
		return NewBlockSpecifierFromString(value.(string))
	case map[string]interface{}:
		return NewBlockSpecifierFromMap(value.(map[string]interface{}))
	default:
		return nil, errors.New(
			"the input value must be an EIP-1898 compatible object (string or map)")
	}
}

func MustBlockSpecifier(value interface{}) *BlockSpecifier {
	b, err := NewBlockSpecifier(value)
	if err != nil {
		panic(err)
	}
	return b
}

func (b *BlockSpecifier) UnmarshalJSON(data []byte) error {
	var v interface{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	_b, err := NewBlockSpecifier(v)
	if err != nil {
		return err
	}

	*b = *_b
	return nil
}

func (b *BlockSpecifier) MarshalJSON() ([]byte, error) {
	if b.Tag != nil {
		// "earliest"
		return json.Marshal(b.Tag)
	}
	if b.Number != nil {
		if b.Raw {
			// "0x0"
			return json.Marshal(b.Number.String())
		}
		// { "blockNumber": "0x0" }
		return json.Marshal(map[string]interface{}{
			"blockNumber": b.Number.String(),
		})
	}
	if b.Hash != nil {
		// { "blockHash": "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3", "requireCanonical": false }
		// a blockhash gets marshalled as object regardless of the .Raw attribute
		return json.Marshal(map[string]interface{}{
			"blockHash":        b.Hash.String(),
			"requireCanonical": b.RequireCanonical,
		})
	}
	return nil, errors.New("cannot marshal an empty block specifier")
}
