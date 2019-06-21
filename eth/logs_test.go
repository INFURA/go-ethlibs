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
			Message: "null params should parse",
			Payload: `{"address":null, "topics":null, "blockHash": null, "fromBlock": null, "toBlock": null}`,
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

func TestLog_UnmarshalJSON(t *testing.T) {
	type TestCase struct {
		Message  string
		Payload  string
		Expected eth.Log
	}

	tests := []TestCase{
		{
			Message: "parity result should include extra fields",
			Payload: `{"address":"0x8b406b4708a45f115347fc2d020735196f994c5f","blockHash":"0x2cdf35a15eaab70f694b1ef15b6375793848336e00b76b0551082b1fb6130ccd","blockNumber":"0xa4c6b2","data":"0x000000000000000000000000000000000000000000000000000000005cbde7cc","logIndex":"0x0","removed":false,"topics":["0x7493e1cf8d9c2822f00f7838f98c6e5f3def8e7251951bc4c1e37588229510cf"],"transactionHash":"0x9cd71724c1bad4c8e09a52b5bc1d8f037d5c08f4b78626236110ce5e6e1e8cfb","transactionIndex":"0xa","transactionLogIndex":"0x0","type":"mined"}`,
			Expected: eth.Log{
				Address:     *eth.MustAddress("0x8b406b4708a45f115347fc2d020735196f994c5f"),
				BlockHash:   eth.MustHash("0x2cdf35a15eaab70f694b1ef15b6375793848336e00b76b0551082b1fb6130ccd"),
				BlockNumber: eth.MustQuantity("0xa4c6b2"),
				Data:        *eth.MustData("0x000000000000000000000000000000000000000000000000000000005cbde7cc"),
				LogIndex:    eth.MustQuantity("0x0"),
				Removed:     false,
				Topics:      []eth.Topic{*eth.MustTopic("0x7493e1cf8d9c2822f00f7838f98c6e5f3def8e7251951bc4c1e37588229510cf")},
				TxHash:      eth.MustHash("0x9cd71724c1bad4c8e09a52b5bc1d8f037d5c08f4b78626236110ce5e6e1e8cfb"),
				TxIndex:     eth.MustQuantity("0xa"),
				TxLogIndex:  eth.MustQuantity("0x0"),
				Type:        eth.OptionalString("mined"),
			},
		},
		{
			Message: "geth result should not include parity extra fields",
			Payload: `{"address":"0x7ef66b77759e12caf3ddb3e4aff524e577c59d8d","blockHash":"0x34b05565741a73563cbc331342e65da9c69f84666c4e24e514e3834769e0a21e","blockNumber":"0x71321","data":"0xa7ecdb0832bb205ff5f337a9c91843754031d09fc7d9aeb4b044d92d908fd959","logIndex":"0x0","removed":false,"topics":["0x8a22ee899102a366ac8ad0495127319cb1ff2403cfae855f83a89cda1266674d","0x000000000000000000000000000000000000000000000000000000000000002a","0x0000000000000000000000000000000000000000000000000000000000a4c6f6"],"transactionHash":"0x139237652bf69398ec28f776431eb383b5d5748a441c91a8a3be8142aac027fc","transactionIndex":"0x0"}`,
			Expected: eth.Log{
				Address:     *eth.MustAddress("0x7ef66b77759e12caf3ddb3e4aff524e577c59d8d"),
				BlockHash:   eth.MustHash("0x34b05565741a73563cbc331342e65da9c69f84666c4e24e514e3834769e0a21e"),
				BlockNumber: eth.MustQuantity("0x71321"),
				Data:        *eth.MustData("0xa7ecdb0832bb205ff5f337a9c91843754031d09fc7d9aeb4b044d92d908fd959"),
				LogIndex:    eth.MustQuantity("0x0"),
				Removed:     false,
				Topics: []eth.Topic{
					*eth.MustTopic("0x8a22ee899102a366ac8ad0495127319cb1ff2403cfae855f83a89cda1266674d"),
					*eth.MustTopic("0x000000000000000000000000000000000000000000000000000000000000002a"),
					*eth.MustTopic("0x0000000000000000000000000000000000000000000000000000000000a4c6f6"),
				},
				TxHash:  eth.MustHash("0x139237652bf69398ec28f776431eb383b5d5748a441c91a8a3be8142aac027fc"),
				TxIndex: eth.MustQuantity("0x0"),
			},
		},
	}

	for _, test := range tests {
		actual := eth.Log{}
		err := json.Unmarshal([]byte(test.Payload), &actual)
		require.NoError(t, err, test.Message)
		require.Equal(t, test.Expected, actual, test.Message)
	}
}
