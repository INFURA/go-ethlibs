package eth_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ConsenSys/go-ethlibs/eth"
)

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
