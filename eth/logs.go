package eth

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/INFURA/go-ethlibs/rlp"
)

type Log struct {
	Removed     bool      `json:"removed"`
	LogIndex    *Quantity `json:"logIndex"`
	TxIndex     *Quantity `json:"transactionIndex"`
	TxHash      *Hash     `json:"transactionHash"`
	BlockHash   *Hash     `json:"blockHash"`
	BlockNumber *Quantity `json:"blockNumber"`
	Address     Address   `json:"address"`
	Data        Data      `json:"data"`
	Topics      []Topic   `json:"topics"`

	// Parity-specific fields
	TxLogIndex *Quantity `json:"transactionLogIndex,omitempty"`
	Type       *string   `json:"type,omitempty"`
}

func (l *Log) RLP() rlp.Value {
	topics := make([]rlp.Value, len(l.Topics))
	for i := range l.Topics {
		topics[i] = l.Topics[i].RLP()
	}
	return rlp.Value{
		List: []rlp.Value{
			l.Address.RLP(),
			{List: topics},
			l.Data.RLP(),
		},
	}
}

type addrOrArray []Address

func (a *addrOrArray) UnmarshalJSON(data []byte) error {
	str := ""
	err := json.Unmarshal(data, &str)

	// If successfully unmarshaled to non-default value, then data represented a single address
	if err == nil && str != "" {
		addr, err := NewAddress(str)
		if err != nil {
			return err
		}
		*a = []Address{*addr}
		return nil
	}

	// otherwise it should be unmarshaled as an array
	// usually log filters specify at most a single Address, so let's
	// start with a capacity of one
	arr := make([]Address, 0, 1)
	err = json.Unmarshal(data, &arr)
	if err == nil {
		*a = arr
		return nil
	}

	return err
}

func (a *addrOrArray) Array() []Address {
	return []Address(*a)
}

type topicOrArray []Topic

func (t *topicOrArray) UnmarshalJSON(data []byte) error {
	null := []byte("null")
	if bytes.Equal(data, null) {
		*t = make([]Topic, 0, 4)
		return nil
	}

	str := ""
	err := json.Unmarshal(data, &str)
	if err == nil {
		topic, err := NewTopic(str)
		if err != nil {
			return err
		}
		*t = []Topic{*topic}
		return nil
	}

	arr := make([]Topic, 0, 4)
	err = json.Unmarshal(data, &arr)
	if err == nil {
		*t = arr
		return nil
	}

	return err
}

func (t *topicOrArray) Array() []Topic {
	return []Topic(*t)
}

type topics [][]Topic

func (t *topics) Array() [][]Topic {
	return [][]Topic(*t)
}

func (t *topics) UnmarshalJSON(data []byte) error {
	/*
			From: https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_newfilter
			A note on specifying topic filters:

			Topics are order-dependent. A transaction with a log with topics [A, B] will be matched by the following topic filters:

			[] "anything"
			[A] "A in first position (and anything after)"
			[null, B] "anything in first position AND B in second position (and anything after)"
			[A, B] "A in first position AND B in second position (and anything after)"
			[[A, B], [A, B]] "(A OR B) in first position AND (A OR B) in second position (and anything after)"

		Since GoLang equates empty slices and nil, we will use an empty slice instead of null for the in-memory representation.
		Thus, each entry in the input array is either a string or an array
	*/
	var parser []topicOrArray
	err := json.Unmarshal(data, &parser)
	if err != nil {
		return err
	}

	var out [][]Topic
	for i := range parser {
		src := parser[i]
		dst := make([]Topic, 0, len(src))
		for j := range src {
			dst = append(dst, src[j])
		}

		out = append(out, dst)
	}

	// remove any extraneous empty topics from the right-hand side of the array
	for i := len(out) - 1; i >= 0; i-- {
		if len(out[i]) > 0 {
			break
		}

		out = out[:i]
	}

	if len(out) == 0 {
		*t = nil
		return nil
	}

	if len(out) > 4 {
		return errors.New("only up to four topic slots may be specified")
	}

	*t = out
	return nil
}
