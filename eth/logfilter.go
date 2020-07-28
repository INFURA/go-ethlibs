package eth

import "encoding/json"

type LogFilter struct {
	FromBlock *BlockNumberOrTag `json:"fromBlock,omitempty"`
	ToBlock   *BlockNumberOrTag `json:"toBlock,omitempty"`
	BlockHash *Hash             `json:"blockHash,omitempty"`
	Address   []Address         `json:"address,omitempty"`
	Topics    [][]Topic         `json:"topics,omitempty"`
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
