package node

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/INFURA/go-ethlibs/eth"
	"github.com/INFURA/go-ethlibs/jsonrpc"
	"github.com/pkg/errors"
	"net/url"
)

var (
	ErrBlockNotFound       = errors.New("block not found")
	ErrTransactionNotFound = errors.New("transaction not found")
)

var _ Client = (*client)(nil)

func NewClient(ctx context.Context, rawURL string) (Client, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse url")
	}

	var transport transport

	switch parsedURL.Scheme {
	case "http", "https":
		transport, err = newHTTPTransport(ctx, parsedURL)
	case "wss", "ws":
		transport, err = newWebsocketTransport(ctx, parsedURL)
	default:
		transport, err = newIPCTransport(ctx, parsedURL)
	}

	if err != nil {
		return nil, errors.Wrap(err, "could not create client transport")
	}

	return &client{
		transport: transport,
		rawURL:    rawURL,
	}, nil
}

func NewCustomClient(requester Requester, subscriber Subscriber) (Client, error) {
	t, err := newCustomTransport(requester, subscriber)
	if err != nil {
		return nil, errors.Wrap(err, "could not create custom transport")
	}

	return &client{
		transport: t,
		rawURL:    "",
	}, nil
}

type transport interface {
	Requester
	Subscriber

	IsBidirectional() bool
}

type client struct {
	transport transport
	rawURL    string
}

func (c *client) Request(ctx context.Context, r *jsonrpc.Request) (*jsonrpc.RawResponse, error) {
	return c.transport.Request(ctx, r)
}

func (c *client) Subscribe(ctx context.Context, r *jsonrpc.Request) (Subscription, error) {
	return c.transport.Subscribe(ctx, r)
}

func (c *client) IsBidirectional() bool {
	return c.transport.IsBidirectional()
}

func (c *client) URL() string {
	return c.rawURL
}

func (c *client) BlockNumber(ctx context.Context, options ...RequestOption) (uint64, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_blockNumber",
	}

	applyOptions(&request, options)
	response, err := c.Request(ctx, &request)
	if err != nil {
		return 0, err
	}

	if response.Error != nil {
		return 0, errors.New(string(*response.Error))
	}

	q := eth.Quantity{}
	err = json.Unmarshal(response.Result, &q)
	if err != nil {
		return 0, errors.Wrap(err, "could not decode result")
	}

	return q.UInt64(), nil
}

func (c *client) BlockByNumber(ctx context.Context, number uint64, full bool, options ...RequestOption) (*eth.Block, error) {
	n := eth.QuantityFromUInt64(number)

	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getBlockByNumber",
		Params: jsonrpc.MustParams(&n, full),
	}

	applyOptions(&request, options)
	response, err := c.Request(ctx, &request)
	if err != nil {
		return nil, errors.Wrap(err, "could not make request")
	}

	return c.parseBlockResponse(response)
}

func (c *client) BlockByNumberOrTag(ctx context.Context, numberOrTag eth.BlockNumberOrTag, full bool, options ...RequestOption) (*eth.Block, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getBlockByNumber",
		Params: jsonrpc.MustParams(&numberOrTag, full),
	}

	applyOptions(&request, options)
	response, err := c.Request(ctx, &request)
	if err != nil {
		return nil, errors.Wrap(err, "could not make request")
	}

	return c.parseBlockResponse(response)
}

func (c *client) BlockByHash(ctx context.Context, hash string, full bool, options ...RequestOption) (*eth.Block, error) {
	h, err := eth.NewHash(hash)
	if err != nil {
		return nil, errors.Wrap(err, "invalid hash")
	}

	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getBlockByHash",
		Params: jsonrpc.MustParams(h, full),
	}

	applyOptions(&request, options)
	response, err := c.Request(ctx, &request)
	if err != nil {
		return nil, errors.Wrap(err, "could not make request")
	}

	return c.parseBlockResponse(response)
}

func (c *client) parseBlockResponse(response *jsonrpc.RawResponse) (*eth.Block, error) {
	if response.Error != nil {
		return nil, errors.New(string(*response.Error))
	}

	if len(response.Result) == 0 || bytes.Equal(response.Result, json.RawMessage(`null`)) {
		return nil, ErrBlockNotFound
	}

	// log.Printf("[SPAM] Result: %s", string(response.Result))

	block := eth.Block{}
	err := json.Unmarshal(response.Result, &block)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode block result")
	}

	return &block, nil
}

func (c *client) TransactionReceipt(ctx context.Context, hash string, options ...RequestOption) (*eth.TransactionReceipt, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getTransactionReceipt",
		Params: jsonrpc.MustParams(hash),
	}

	applyOptions(&request, options)
	response, err := c.Request(ctx, &request)
	if err != nil {
		return nil, errors.Wrap(err, "could not make request")
	}

	if response.Error != nil {
		return nil, errors.New(string(*response.Error))
	}

	if len(response.Result) == 0 || bytes.Equal(response.Result, json.RawMessage(`null`)) {
		// Then the transaction isn't recognized
		return nil, errors.Errorf("receipt for transaction %s not found", hash)
	}

	receipt := eth.TransactionReceipt{}
	err = json.Unmarshal(response.Result, &receipt)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal result")
	}

	return &receipt, nil
}

func (c *client) Logs(ctx context.Context, filter eth.LogFilter, options ...RequestOption) ([]eth.Log, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getLogs",
		Params: jsonrpc.MustParams(filter),
	}

	applyOptions(&request, options)
	response, err := c.Request(ctx, &request)
	if err != nil {
		return nil, errors.Wrap(err, "could not make request")
	}

	if response.Error != nil {
		return nil, errors.New(string(*response.Error))
	}

	_logs := make([]eth.Log, 0)
	err = json.Unmarshal(response.Result, &_logs)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal result")
	}

	return _logs, nil
}

func (c *client) TransactionByHash(ctx context.Context, hash string, options ...RequestOption) (*eth.Transaction, error) {
	h, err := eth.NewHash(hash)
	if err != nil {
		return nil, errors.Wrap(err, "invalid hash")
	}

	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getTransactionByHash",
		Params: jsonrpc.MustParams(h),
	}

	applyOptions(&request, options)
	response, err := c.Request(ctx, &request)
	if err != nil {
		return nil, errors.Wrap(err, "could not make transaction by hash request")
	}

	if len(response.Result) == 0 || bytes.Equal(response.Result, json.RawMessage(`null`)) {
		// Then the transaction isn't recognized
		return nil, ErrTransactionNotFound
	}

	tx := eth.Transaction{}
	err = tx.UnmarshalJSON(response.Result)
	return &tx, err
}

func (c *client) SubscribeNewHeads(ctx context.Context, options ...RequestOption) (Subscription, error) {
	request := jsonrpc.Request{
		JSONRPC: "2.0",
		ID:      jsonrpc.ID{Str: "test", IsString: true},
		Method:  "eth_subscribe",
		Params:  jsonrpc.MustParams("newHeads"),
	}

	applyOptions(&request, options)
	return c.Subscribe(ctx, &request)
}

func (c *client) SubscribeNewPendingTransactions(ctx context.Context, options ...RequestOption) (Subscription, error) {
	request := jsonrpc.Request{
		JSONRPC: "2.0",
		ID:      jsonrpc.ID{Str: "pending", IsString: true},
		Method:  "eth_subscribe",
		Params:  jsonrpc.MustParams("newPendingTransactions"),
	}

	applyOptions(&request, options)
	return c.Subscribe(ctx, &request)
}

func applyOptions(request *jsonrpc.Request, options []RequestOption) {
	for _, f := range options {
		f(request)
	}
}
