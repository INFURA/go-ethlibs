package jsonrpc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalGoodNotifications(t *testing.T) {
	type TestCase struct {
		Description  string
		Raw          string
		Notification Notification
	}

	testCases := []TestCase{
		{
			Description: "Notification result from Parity",
			Raw:         `{"jsonrpc":"2.0","method":"parity_subscription","params":{"result":"0x3342d6","subscription":"0x0c2f1dc472de1be0"}}`,
			Notification: Notification{
				JSONRPC: "2.0",
				Method:  "parity_subscription",
				Params:  NotificationParams(`{"result":"0x3342d6","subscription":"0x0c2f1dc472de1be0"}`),
			},
		},
		{
			Description: "Notification result from Geth",
			Raw:         `{"jsonrpc":"2.0","method":"eth_subscription","params":{"subscription":"0x3eb3487232a1bf601f92757e0a5d0b18","result":{"parentHash":"0x28f2668a84038a5b07d13564b7b11421c7ca74867f80a06c8bf429057c5000dd","truncated":"..."}}}`,
			Notification: Notification{
				JSONRPC: "2.0",
				Method:  "eth_subscription",
				Params:  NotificationParams(`{"subscription":"0x3eb3487232a1bf601f92757e0a5d0b18","result":{"parentHash":"0x28f2668a84038a5b07d13564b7b11421c7ca74867f80a06c8bf429057c5000dd","truncated":"..."}}`),
			},
		},
	}

	for _, testCase := range testCases {
		parsed := Notification{}

		err := json.Unmarshal([]byte(testCase.Raw), &parsed)
		if assert.Nil(t, err, "Got err '%v' parsing '%s'", err, testCase.Raw) == false {
			continue
		}

		assert.Equal(t, testCase.Notification, parsed)
	}
}
