package eth_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/INFURA/ethereum-interaction/pkg/eth"
)

func TestLogFilterParsing(t *testing.T) {
	type TestCase struct {
		Message  string
		Payload  string
		Expected eth.LogFilter
	}

	tests := []TestCase{
		{
			Message: "empty params should parse",
			Payload: `{}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "single address as a string should be supported",
			Payload: `{"address": "0xcdbf1d1c64faad6d8484fd3dd5be2b4ea57f5f4c"}`,
			Expected: eth.LogFilter{
				Address:   []eth.Address{*eth.MustAddress("0xcdbf1d1c64faad6d8484fd3dd5be2b4ea57f5f4c")},
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "single address as an array should be supported",
			Payload: `{"address": ["0xcdbf1d1c64faad6d8484fd3dd5be2b4ea57f5f4c"]}`,
			Expected: eth.LogFilter{
				Address:   []eth.Address{*eth.MustAddress("0xcdbf1d1c64faad6d8484fd3dd5be2b4ea57f5f4c")},
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "multiple addresses should be supported",
			Payload: `{"address": ["0xcdbf1d1c64faad6d8484fd3dd5be2b4ea57f5f4c", "0x5717adf502fd8830456bd5dc26801a4db394e6b2"]}`,
			Expected: eth.LogFilter{
				Address: []eth.Address{
					*eth.MustAddress("0xcdbf1d1c64faad6d8484fd3dd5be2b4ea57f5f4c"),
					*eth.MustAddress("0x5717adf502fd8830456bd5dc26801a4db394e6b2"),
				},
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "single topic should be supported",
			Payload: `{"topics": ["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"]}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics: [][]eth.Topic{
					{eth.Topic("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
				},
			},
		},
		{
			Message: "multiple topics should be supported",
			Payload: `{"topics": ["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", "0x000000000000000000000000896dd350806eba53dfa9778c4698224e8ede2c41"]}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics: [][]eth.Topic{
					{eth.Topic("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
					{eth.Topic("0x000000000000000000000000896dd350806eba53dfa9778c4698224e8ede2c41")},
				},
			},
		},
		{
			Message: "extraneous null topics should be filtered",
			Payload: `{"topics": [null]}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "extraneous null topics should be filtered",
			Payload: `{"topics": ["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", null]}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics: [][]eth.Topic{
					{eth.Topic("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
				},
			},
		},
		{
			Message: "null topics should be supported",
			Payload: `{"topics": [null, "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"]}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics: [][]eth.Topic{
					{},
					{eth.Topic("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
				},
			},
		},
		{
			Message: "ORed topics should be supported",
			Payload: `{"topics": [["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", "0x9577941d28fff863bfbee4694a6a4a56fb09e169619189d2eaa750b5b4819995"]]}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: nil,
				Topics: [][]eth.Topic{
					{
						eth.Topic("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
						eth.Topic("0x9577941d28fff863bfbee4694a6a4a56fb09e169619189d2eaa750b5b4819995"),
					},
				},
			},
		},
		{
			Message: "blockHash must be supported",
			Payload: `{"blockHash":"0xb509a2149556380fbff167f2fdfad07cf9cfe8eb605e83298683008f46f419b5"}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   nil,
				BlockHash: eth.MustHash("0xb509a2149556380fbff167f2fdfad07cf9cfe8eb605e83298683008f46f419b5"),
				Topics:    nil,
			},
		},
		{
			Message: "fromBlock must be supported",
			Payload: `{"fromBlock":"0x1234"}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: eth.MustBlockNumberOrTag("0x1234"),
				ToBlock:   nil,
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "toBlock must be supported",
			Payload: `{"toBlock":"0x1234"}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: nil,
				ToBlock:   eth.MustBlockNumberOrTag("0x1234"),
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "block tags must be supported, tags matching defaults should still be present",
			Payload: `{"fromBlock":"latest", "toBlock": "latest"}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: eth.MustBlockNumberOrTag("latest"),
				ToBlock:   eth.MustBlockNumberOrTag("latest"),
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "block tags must be supported",
			Payload: `{"fromBlock":"earliest", "toBlock": "earliest"}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: eth.MustBlockNumberOrTag("earliest"),
				ToBlock:   eth.MustBlockNumberOrTag("earliest"),
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "block tags must be supported",
			Payload: `{"fromBlock":"pending", "toBlock": "pending"}`,
			Expected: eth.LogFilter{
				Address:   nil,
				FromBlock: eth.MustBlockNumberOrTag(eth.TagPending),
				ToBlock:   eth.MustBlockNumberOrTag(eth.TagPending),
				BlockHash: nil,
				Topics:    nil,
			},
		},
		{
			Message: "complex example",
			Payload: `{
				"fromBlock":"0x1234", 
				"toBlock": "latest", 
				"blockHash":"0xb509a2149556380fbff167f2fdfad07cf9cfe8eb605e83298683008f46f419b5", 
				"address": ["0xcdbf1d1c64faad6d8484fd3dd5be2b4ea57f5f4c", "0x5717adf502fd8830456bd5dc26801a4db394e6b2"],
				"topics": [["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", "0x9577941d28fff863bfbee4694a6a4a56fb09e169619189d2eaa750b5b4819995"], "0x000000000000000000000000896dd350806eba53dfa9778c4698224e8ede2c41"]
			}`,
			Expected: eth.LogFilter{
				Address: []eth.Address{
					*eth.MustAddress("0xcdbf1d1c64faad6d8484fd3dd5be2b4ea57f5f4c"),
					*eth.MustAddress("0x5717adf502fd8830456bd5dc26801a4db394e6b2"),
				},
				FromBlock: eth.MustBlockNumberOrTag("0x1234"),
				ToBlock:   eth.MustBlockNumberOrTag("latest"),
				BlockHash: eth.MustHash("0xb509a2149556380fbff167f2fdfad07cf9cfe8eb605e83298683008f46f419b5"),
				Topics: [][]eth.Topic{
					{
						eth.Topic("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
						eth.Topic("0x9577941d28fff863bfbee4694a6a4a56fb09e169619189d2eaa750b5b4819995"),
					},
					{eth.Topic("0x000000000000000000000000896dd350806eba53dfa9778c4698224e8ede2c41")},
				},
			},
		},
	}

	for _, test := range tests {
		actual := eth.LogFilter{}
		err := json.Unmarshal([]byte(test.Payload), &actual)
		require.NoError(t, err, test.Message)
		require.Equal(t, test.Expected, actual, test.Message)
	}

}
