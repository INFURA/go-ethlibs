package jsonrpc

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleRequestServer(t *testing.T) {
	handled := false

	server := httptest.NewServer(RequestHandlerFunc(func(ctx RequestContext, r *Request) (interface{}, *Error) {
		if r.Method != "eth_blockNumber" {
			handled = false
			return nil, MethodNotSupported(r)
		}

		handled = true
		assert.Equal(t, "application/json", ctx.HTTPRequest().Header.Get("Content-Type"))

		b, _ := json.Marshal(r)
		assert.Equal(t, json.RawMessage(b), ctx.RawJSON())
		return "0x123456", nil
	}))

	client := http.Client{}

	req := MustRequest(1, "eth_blockNumber")
	b, _ := json.Marshal(&req)

	httpReq, _ := http.NewRequest(http.MethodPost, server.URL, bytes.NewReader(b))
	httpReq.Header.Set("Content-Type", "application/json")

	res, _ := client.Do(httpReq)
	assert.True(t, handled, "not handled when expected to be")
	assert.Equal(t, http.StatusOK, res.StatusCode, "response should be code 200")

	// double check that the outputted JSON contains "result" but not "error"
	b, _ = ioutil.ReadAll(res.Body)
	assert.Contains(t, string(b), `"result"`, "response should contain result element")
	assert.NotContains(t, string(b), `"error"`, "response should not contain error element")

	req.Method = "foo_notSupported"
	b, _ = json.Marshal(&req)

	httpReq, _ = http.NewRequest(http.MethodPost, server.URL, bytes.NewReader(b))
	httpReq.Header.Set("Content-Type", "application/json")

	res, _ = client.Do(httpReq)
	assert.False(t, handled, "handled when expected not to be")
	assert.Equal(t, http.StatusOK, res.StatusCode, "response should be code 200 for errors too")

	// double check that the outputted JSON contains "error" but not "result"
	b, _ = ioutil.ReadAll(res.Body)
	assert.Contains(t, string(b), `"error"`, "response should contain error element")
	assert.NotContains(t, string(b), `"result"`, "response should not contain result element")
}

func ExampleRequestHandlerFunc() {
	http.Handle("/", RequestHandlerFunc(func(ctx RequestContext, r *Request) (interface{}, *Error) {
		if r.Method != "eth_blockNumber" {
			return nil, MethodNotSupported(r)
		}

		// if the underlying HTTP Request object is required, it is accessible from the context
		url := ctx.HTTPRequest().URL
		if url.Host != "mainnet.infura.io" {
			return nil, InvalidInput("wrong host")
		}

		return "0x123456", nil
	}))
}
