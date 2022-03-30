package jsonrpc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalGoodRequests(t *testing.T) {

	type TestCase struct {
		Description string
		Raw         string
		Request     *Request
	}

	testCases := []TestCase{
		{
			Description: "String ID with No Params",
			Raw:         `{"method":"eth_blockNumber","id":"27a5fbbcaa23c1dcca4deb04f1501efb","jsonrpc":"2.0"}`,
			Request: &Request{
				ID: ID{
					Str:      "27a5fbbcaa23c1dcca4deb04f1501efb",
					IsString: true,
				},
				Method: "eth_blockNumber",
				Params: nil,
			},
		},
		{
			Description: "Int ID with Empty Params",
			Raw:         `{"method":"eth_blockNumber","id": 42,"jsonrpc":"2.0", "params":[]}`,
			Request: &Request{
				ID: ID{
					Num: 42,
				},
				Method: "eth_blockNumber",
				Params: Params{},
			},
		},
		{
			Description: "Int ID with Null Params",
			Raw:         `{"method":"eth_blockNumber","id": 42,"jsonrpc":"2.0", "params":null}`,
			Request: &Request{
				ID: ID{
					Num: 42,
				},
				Method: "eth_blockNumber",
				Params: nil,
			},
		},
		{
			Description: "Int ID with Empty Params, but no JSONRPC version",
			Raw:         `{"method":"eth_blockNumber","id": 42,"params":[]}`,
			Request: &Request{
				ID: ID{
					Num: 42,
				},
				Method: "eth_blockNumber",
				Params: Params{},
			},
		},
		{
			Description: "Int ID with Single Param",
			Raw:         `{"method":"eth_blockNumber","id": 42,"jsonrpc":"2.0", "params":["string"]}`,
			Request: &Request{
				ID: ID{
					Num: 42,
				},
				Method: "eth_blockNumber",
				Params: MustParams("string"),
			},
		},
		{
			Description: "Int ID with Single Object Param",
			Raw:         `{"method":"eth_blockNumber","id": 42,"jsonrpc":"2.0", "params":[{"foo":"bar"}]}`,
			Request: &Request{
				ID: ID{
					Num: 42,
				},
				Method: "eth_blockNumber",
				Params: MustParams(map[string]string{"foo": "bar"}),
			},
		},
		{
			Description: "",
			Raw:         `{"id":1839673506133526,"jsonrpc":"2.0","params":["0x599784",true],"method":"eth_getBlockByNumber"}`,
			Request: &Request{
				ID: ID{
					Num: 1839673506133526,
				},
				Method: "eth_getBlockByNumber",
				Params: MustParams("0x599784", true),
			},
		},
	}

	for _, testCase := range testCases {
		parsed := Request{}

		err := json.Unmarshal([]byte(testCase.Raw), &parsed)
		if assert.Nil(t, err, "Got err '%v' parsing '%s'", err, testCase.Raw) == false {
			continue
		}

		assert.Equal(t, testCase.Request.ID, parsed.ID)
		assert.Equal(t, testCase.Request.Method, parsed.Method)
		assert.Equal(t, testCase.Request.Params, parsed.Params)

		//remarshal without error
		_, err = json.Marshal(&parsed)
		assert.Nil(t, err, "Got err '%v' re-Marshaling parsed JSON")
	}
}

func TestUnmarshalBadRequests(t *testing.T) {
	type TestCase struct {
		Description string
		Raw         string
	}

	testCases := []TestCase{
		{
			Description: "Non-object JSON",
			Raw:         `42`,
		},
		{
			Description: "Empty string",
			Raw:         ``,
		},
		{
			Description: "Empty object, should be not be decoded as a Request, but could be a Notification object",
			Raw:         `{}`,
		},
		{
			Description: "Notification result with no ID, should not parse",
			Raw:         `{"jsonrpc":"2.0","method":"parity_subscription","params":{"result":"0x3342d6","subscription":"0x0c2f1dc472de1be0"}}`,
		},
	}

	for _, testCase := range testCases {
		parsed := Request{}
		err := json.Unmarshal([]byte(testCase.Raw), &parsed)
		if assert.NotNil(t, err, "Expected err but got %v parsing: %s", err, testCase.Raw) == false {
			continue
		}
	}
}

func TestRemarshalRequestWithNetworks(t *testing.T) {

	type TestCase struct {
		Description string
		Raw         string
		Request     *Request
	}

	testCases := []TestCase{
		{
			Description: "String ID with No Params",
			Raw:         `{"method":"eth_blockNumber","id":"27a5fbbcaa23c1dcca4deb04f1501efb","jsonrpc":"2.0"}`,
			Request: &Request{
				ID: ID{
					Str:      "27a5fbbcaa23c1dcca4deb04f1501efb",
					IsString: true,
				},
				Method: "eth_blockNumber",
				Params: nil,
			},
		},
		{
			Description: "Int ID with Empty Params",
			Raw:         `{"method":"eth_blockNumber","id": 42,"jsonrpc":"2.0", "params":[]}`,
			Request: &Request{
				ID: ID{
					Num: 42,
				},
				Method: "eth_blockNumber",
				Params: Params{},
			},
		},
		{
			Description: "Int ID with Null Params",
			Raw:         `{"method":"eth_blockNumber","id": 42,"jsonrpc":"2.0", "params":null}`,
			Request: &Request{
				ID: ID{
					Num: 42,
				},
				Method: "eth_blockNumber",
				Params: nil,
			},
		},
		{
			Description: "Int ID with Empty Params, but no JSONRPC version",
			Raw:         `{"method":"eth_blockNumber","id": 42,"params":[]}`,
			Request: &Request{
				ID: ID{
					Num: 42,
				},
				Method: "eth_blockNumber",
				Params: Params{},
			},
		},
		{
			Description: "Int ID with Single Param",
			Raw:         `{"method":"eth_blockNumber","id": 42,"jsonrpc":"2.0", "params":["string"]}`,
			Request: &Request{
				ID: ID{
					Num: 42,
				},
				Method: "eth_blockNumber",
				Params: MustParams("string"),
			},
		},
		{
			Description: "Int ID with Single Object Param",
			Raw:         `{"method":"eth_blockNumber","id": 42,"jsonrpc":"2.0", "params":[{"foo":"bar"}]}`,
			Request: &Request{
				ID: ID{
					Num: 42,
				},
				Method: "eth_blockNumber",
				Params: MustParams(map[string]string{"foo": "bar"}),
			},
		},
		{
			Description: "",
			Raw:         `{"id":1839673506133526,"jsonrpc":"2.0","params":["0x599784",true],"method":"eth_getBlockByNumber"}`,
			Request: &Request{
				ID: ID{
					Num: 1839673506133526,
				},
				Method: "eth_getBlockByNumber",
				Params: MustParams("0x599784", true),
			},
		},
	}

	for _, testCase := range testCases {
		parsed := Request{}

		err := json.Unmarshal([]byte(testCase.Raw), &parsed)
		if assert.NoError(t, err, "Testcase: %v", testCase) == false {
			continue
		}

		assert.Equal(t, testCase.Request.ID, parsed.ID)
		assert.Equal(t, testCase.Request.Method, parsed.Method)
		assert.Equal(t, testCase.Request.Params, parsed.Params)

		//remarshal as a RequestWithNetwork without error
		network := "mainnet"
		b, err := json.Marshal(RequestWithNetwork{&parsed, network})
		assert.Nil(t, err, "Got err '%v' re-Marshaling parsed JSON")

		var m interface{}
		err = json.Unmarshal(b, &m)
		assert.Nil(t, err, "Got err '%v' unmarshaling re-marshaled JSON")

		assert.Equal(t, network, m.(map[string]interface{})["network"])
	}
}

func TestMarshalGoodRequests(t *testing.T) {

	type TestCase struct {
		Description string
		Expected    string
		Request     *Request
	}

	testCases := []TestCase{
		{
			Description: "String ID with No Params",
			Expected:    `{"method":"eth_blockNumber","params":[],"id":"27a5fbbcaa23c1dcca4deb04f1501efb","jsonrpc":"2.0"}`,
			Request: &Request{
				ID: ID{
					Str:      "27a5fbbcaa23c1dcca4deb04f1501efb",
					IsString: true,
				},
				Method: "eth_blockNumber",
				Params: nil,
			},
		},
		{
			Description: "Int ID with Empty Params",
			Expected:    `{"method":"eth_blockNumber","params":[],"id":42,"jsonrpc":"2.0"}`,
			Request: &Request{
				ID: ID{
					Num: 42,
				},
				Method: "eth_blockNumber",
				Params: Params{},
			},
		},
		{
			Description: "Int ID with Null Params",
			Expected:    `{"method":"eth_blockNumber","params":[],"id":42,"jsonrpc":"2.0"}`,
			Request: &Request{
				ID: ID{
					Num: 42,
				},
				Method: "eth_blockNumber",
				Params: nil,
			},
		},
		{
			Description: "Int ID with Empty Params, but no JSONRPC version",
			Expected:    `{"method":"eth_blockNumber","params":[],"id":42,"jsonrpc":"2.0"}`,
			Request: &Request{
				ID: ID{
					Num: 42,
				},
				Method: "eth_blockNumber",
				Params: Params{},
			},
		},
		{
			Description: "Int ID with Single Param",
			Expected:    `{"method":"eth_blockNumber","params":["string"],"id":42,"jsonrpc":"2.0"}`,
			Request: &Request{
				ID: ID{
					Num: 42,
				},
				Method: "eth_blockNumber",
				Params: MustParams("string"),
			},
		},
		{
			Description: "Int ID with Single Object Param",
			Expected:    `{"method":"eth_blockNumber","params":[{"foo":"bar"}],"id":42,"jsonrpc":"2.0"}`,
			Request: &Request{
				ID: ID{
					Num: 42,
				},
				Method: "eth_blockNumber",
				Params: MustParams(map[string]string{"foo": "bar"}),
			},
		},
		{
			Description: "",
			Expected:    `{"method":"eth_getBlockByNumber","params":["0x599784",true],"id":1839673506133526,"jsonrpc":"2.0"}`,
			Request: &Request{
				ID: ID{
					Num: 1839673506133526,
				},
				Method: "eth_getBlockByNumber",
				Params: MustParams("0x599784", true),
			},
		},
	}

	for _, testCase := range testCases {
		res, err := json.Marshal(testCase.Request)
		if assert.NoError(t, err, "Got err '%v' marshal '%q'", err, testCase.Expected) == false {
			continue
		}

		assert.Equal(t, testCase.Expected, string(res), testCase.Description)

		req := Request{}
		err = json.Unmarshal(res, &req)
		assert.NoError(t, err, "Got err '%v' unmarshal '%s'", err, string(res))
	}
}
