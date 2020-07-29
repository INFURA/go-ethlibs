package eth

import "encoding/json"

type LogFilter struct {
	FromBlock *BlockNumberOrTag `json:"fromBlock,omitempty"`
	ToBlock   *BlockNumberOrTag `json:"toBlock,omitempty"`
	BlockHash *Hash             `json:"blockHash,omitempty"`
	Address   []Address         `json:"address,omitempty"`
	Topics    [][]Topic         `json:"topics,omitempty"`
}

// Matches returns true if this filter matches the passed in log.  It follows the rules expressed at:
// https://eth.wiki/json-rpc/API#a-note-on-specifying-topic-filters
//
// However, note that "tags" for FromBlock and ToBlock are implicitly ignored and treated the same as if
// the parameter was nil.  Callers should replace tags with concrete quantities for block numbers before
// calling this function.
func (f *LogFilter) Matches(l Log) bool {
	matchBlock := func() bool {
		if f.BlockHash != nil && f.BlockHash.String() != l.BlockHash.String() {
			return false
		}

		if f.FromBlock != nil {
			if q, ok := f.FromBlock.Quantity(); ok && l.BlockNumber.UInt64() < q.UInt64() {
				return false
			}
		}

		if f.ToBlock != nil {
			if q, ok := f.ToBlock.Quantity(); ok && l.BlockNumber.UInt64() > q.UInt64() {
				return false
			}
		}

		return true
	}

	matchAddress := func() bool {
		if len(f.Address) == 0 {
			return true
		}

		for i := range f.Address {
			if f.Address[i].String() == l.Address.String() {
				return true
			}
		}

		return false
	}

	matchTopic := func(i int) bool {
		if len(f.Topics[i]) == 0 {
			return true
		}

		if len(l.Topics) <= i {
			// log doesn't have enough topics to possibly match
			return false
		}

		for j := range f.Topics[i] {
			if f.Topics[i][j].String() == l.Topics[i].String() {
				return true
			}
		}

		return false
	}

	matchTopics := func() bool {
		if len(f.Topics) == 0 {
			return true
		}

		for i := range f.Topics {
			if matchTopic(i) == false {
				return false
			}
		}

		return true
	}

	if matchBlock() && matchAddress() && matchTopics() {
		return true
	}

	return false
}

func (f *LogFilter) UnmarshalJSON(data []byte) error {
	// address can be either a single string or an array of strings
	// topics is an array where each item is either a topic or array of topics
	type params struct {
		FromBlock *BlockNumberOrTag `json:"fromBlock,omitempty"`
		ToBlock   *BlockNumberOrTag `json:"toBlock,omitempty"`
		BlockHash *Hash             `json:"blockHash,omitempty"`
		Address   addrOrArray       `json:"address,omitempty"`
		Topics    topics            `json:"topics,omitempty"`
	}

	parser := params{}
	err := json.Unmarshal(data, &parser)
	if err != nil {
		return err
	}

	f.FromBlock = parser.FromBlock
	f.ToBlock = parser.ToBlock
	f.BlockHash = parser.BlockHash
	f.Address = parser.Address.Array()
	f.Topics = parser.Topics.Array()
	return nil
}
