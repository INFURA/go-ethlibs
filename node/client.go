package node

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/INFURA/go-ethlibs/eth"
	"github.com/INFURA/go-ethlibs/jsonrpc"
)

var (
	ErrBlockNotFound       = errors.New("block not found")
	ErrTransactionNotFound = errors.New("transaction not found")
)

var _ Client = (*client)(nil)

func NewClient(ctx context.Context, rawURL string, requestHeader http.Header) (Client, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse url")
	}

	var transport transport

	switch parsedURL.Scheme {
	case "http", "https":
		transport, err = newHTTPTransport(ctx, parsedURL, requestHeader)
	case "wss", "ws":
		transport, err = newWebsocketTransport(ctx, parsedURL, requestHeader)
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

func (c *client) BlockNumber(ctx context.Context) (uint64, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_blockNumber",
	}

	applyContext(ctx, &request)
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

func (c *client) GetTransactionCount(ctx context.Context, address eth.Address, numberOrTag eth.BlockNumberOrTag) (uint64, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getTransactionCount",
		Params: jsonrpc.MustParams(address, &numberOrTag),
	}

	applyContext(ctx, &request)
	response, err := c.Request(ctx, &request)
	if err != nil {
		return 0, errors.Wrap(err, "could not make request")
	}

	if response.Error != nil {
		return 0, errors.New(string(*response.Error))
	}

	q := eth.Quantity{}
	err = json.Unmarshal(response.Result, &q)
	if err != nil {
		return 0, errors.Wrap(err, "could not decode result")
	}

	return q.UInt64(), err
}

func (c *client) NetVersion(ctx context.Context) (string, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "net_version",
		Params: nil,
	}

	applyContext(ctx, &request)
	response, err := c.Request(ctx, &request)
	if err != nil {
		return "", errors.Wrap(err, "could not make request")
	}

	if response.Error != nil {
		return "", errors.New(string(*response.Error))
	}

	version := ""
	err = json.Unmarshal(response.Result, &version)
	if err != nil {
		return "", errors.Wrap(err, "could not decode result")
	}

	return version, nil
}

func (c *client) ChainId(ctx context.Context) (string, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_chainId",
		Params: nil,
	}

	applyContext(ctx, &request)
	response, err := c.Request(ctx, &request)
	if err != nil {
		return "", errors.Wrap(err, "could not make request")
	}

	if response.Error != nil {
		return "", errors.New(string(*response.Error))
	}

	chainId := ""
	err = json.Unmarshal(response.Result, &chainId)
	if err != nil {
		return "", errors.Wrap(err, "could not decode result")
	}

	return chainId, nil
}

func (c *client) BlockByNumber(ctx context.Context, number uint64, full bool) (*eth.Block, error) {
	n := eth.QuantityFromUInt64(number)

	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getBlockByNumber",
		Params: jsonrpc.MustParams(&n, full),
	}

	applyContext(ctx, &request)
	response, err := c.Request(ctx, &request)
	if err != nil {
		return nil, errors.Wrap(err, "could not make request")
	}

	return c.parseBlockResponse(response)
}

func (c *client) EstimateGas(ctx context.Context, msg eth.Transaction) (uint64, error) {
	arg := map[string]interface{}{
		"from":  msg.From,
		"to":    msg.To,
		"value": msg.Value.String(),
	}
	if len(msg.Input) > 0 {
		arg["data"] = msg.Input
	}

	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_estimateGas",
		Params: jsonrpc.MustParams(arg),
	}
	applyContext(ctx, &request)
	response, err := c.Request(ctx, &request)
	if err != nil {
		return 0, errors.Wrap(err, "could not make request")
	}

	if response.Error != nil {
		return 0, errors.New(string(*response.Error))
	}

	q := eth.Quantity{}
	err = json.Unmarshal(response.Result, &q)
	if err != nil {
		return 0, errors.Wrap(err, "could not decode result")
	}
	return q.UInt64(), err
}

func (c *client) SendRawTransaction(ctx context.Context, msg string) (string, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_sendRawTransaction",
		Params: jsonrpc.MustParams(msg),
	}

	applyContext(ctx, &request)
	response, err := c.Request(ctx, &request)
	if err != nil {
		return "", errors.Wrap(err, "could not make request")
	}

	if response.Error != nil {
		return "", errors.New(string(*response.Error))
	}

	txHash := eth.Hash("")
	err = json.Unmarshal(response.Result, &txHash)
	if err != nil {
		return "", errors.Wrap(err, "could not decode result")
	}

	return txHash.String(), nil
}

func (c *client) MaxPriorityFeePerGas(ctx context.Context) (uint64, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_maxPriorityFeePerGas",
		Params: nil,
	}

	applyContext(ctx, &request)
	response, err := c.Request(ctx, &request)
	if err != nil {
		return 0, errors.Wrap(err, "could not make request")
	}

	q := eth.Quantity{}
	err = json.Unmarshal(response.Result, &q)
	if err != nil {
		return 0, errors.Wrap(err, "could not decode result")
	}
	return q.UInt64(), err
}

func (c *client) GasPrice(ctx context.Context) (uint64, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_gasPrice",
		Params: nil,
	}

	applyContext(ctx, &request)
	response, err := c.Request(ctx, &request)
	if err != nil {
		return 0, errors.Wrap(err, "could not make request")
	}

	q := eth.Quantity{}
	err = json.Unmarshal(response.Result, &q)
	if err != nil {
		return 0, errors.Wrap(err, "could not decode result")
	}

	return q.UInt64(), err
}

func (c *client) BlockByNumberOrTag(ctx context.Context, numberOrTag eth.BlockNumberOrTag, full bool) (*eth.Block, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getBlockByNumber",
		Params: jsonrpc.MustParams(&numberOrTag, full),
	}

	applyContext(ctx, &request)
	response, err := c.Request(ctx, &request)
	if err != nil {
		return nil, errors.Wrap(err, "could not make request")
	}

	return c.parseBlockResponse(response)
}

func (c *client) BlockByHash(ctx context.Context, hash string, full bool) (*eth.Block, error) {
	h, err := eth.NewHash(hash)
	if err != nil {
		return nil, errors.Wrap(err, "invalid hash")
	}

	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getBlockByHash",
		Params: jsonrpc.MustParams(h, full),
	}

	applyContext(ctx, &request)
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

func (c *client) TransactionReceipt(ctx context.Context, hash string) (*eth.TransactionReceipt, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getTransactionReceipt",
		Params: jsonrpc.MustParams(hash),
	}

	applyContext(ctx, &request)
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

func (c *client) Logs(ctx context.Context, filter eth.LogFilter) ([]eth.Log, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getLogs",
		Params: jsonrpc.MustParams(filter),
	}

	applyContext(ctx, &request)
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

func (c *client) TransactionByHash(ctx context.Context, hash string) (*eth.Transaction, error) {
	h, err := eth.NewHash(hash)
	if err != nil {
		return nil, errors.Wrap(err, "invalid hash")
	}

	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getTransactionByHash",
		Params: jsonrpc.MustParams(h),
	}

	applyContext(ctx, &request)
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

func (c *client) SubscribeNewHeads(ctx context.Context) (Subscription, error) {
	request := jsonrpc.Request{
		JSONRPC: "2.0",
		ID:      jsonrpc.ID{Str: "newHeads", IsString: true},
		Method:  "eth_subscribe",
		Params:  jsonrpc.MustParams("newHeads"),
	}

	applyContext(ctx, &request)
	return c.Subscribe(ctx, &request)
}

func (c *client) SubscribeNewPendingTransactions(ctx context.Context) (Subscription, error) {
	request := jsonrpc.Request{
		JSONRPC: "2.0",
		ID:      jsonrpc.ID{Str: "pending", IsString: true},
		Method:  "eth_subscribe",
		Params:  jsonrpc.MustParams("newPendingTransactions"),
	}

	applyContext(ctx, &request)
	return c.Subscribe(ctx, &request)
}

func applyContext(ctx context.Context, request *jsonrpc.Request) {
	if id := requestIDFromContext(ctx); id != nil {
		request.ID = *id
	}
}

func (c *client) Call(ctx context.Context, msg eth.Transaction, numberOrTag eth.BlockNumberOrTag) (string, error) {
	arg := map[string]interface{}{
		"from":  msg.From,
		"to":    msg.To,
		"value": msg.Value.String(),
	}
	if len(msg.Input) > 0 {
		arg["data"] = msg.Input
	}

	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_call",
		Params: jsonrpc.MustParams(arg, &numberOrTag),
	}

	applyContext(ctx, &request)
	response, err := c.Request(ctx, &request)
	if err != nil {
		return "", errors.Wrap(err, "could not make request")
	}
	if response.Error != nil {
		return "", errors.New(string(*response.Error))
	}

	txHash := eth.Hash("")
	err = json.Unmarshal(response.Result, &txHash)
	if err != nil {
		return "", errors.Wrap(err, "could not decode result")
	}

	return txHash.String(), nil
}

func (c *client) GetAccounts(ctx context.Context) ([]eth.Address, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_accounts",
		Params: nil,
	}

	applyContext(ctx, &request)
	response, err := c.Request(ctx, &request)
	if err != nil {
		return nil, errors.Wrap(err, "could not make request")
	}
	if response.Error != nil {
		return nil, errors.New(string(*response.Error))
	}

	accountList := make([]eth.Address, 0)
	err = json.Unmarshal(response.Result, &accountList)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode result")
	}

	return accountList, nil
}
