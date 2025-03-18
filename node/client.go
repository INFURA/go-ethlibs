// Package node 提供以太坊节点连接和通信功能
// 包含了客户端初始化、区块查询、交易处理和事件订阅等核心功能
// 支持HTTP、WebSocket和IPC三种连接方式
package node

import (
	"bytes"
	"context"
	"encoding/json"
	"net/url"

	"github.com/pkg/errors"

	"github.com/INFURA/go-ethlibs/eth"
	"github.com/INFURA/go-ethlibs/jsonrpc"
)

// 定义常见错误类型
var (
	// ErrBlockNotFound 表示请求的区块未找到
	// 当查询不存在的区块时返回此错误
	ErrBlockNotFound = errors.New("block not found")
	// ErrTransactionNotFound 表示请求的交易未找到
	// 当查询不存在的交易时返回此错误
	ErrTransactionNotFound = errors.New("transaction not found")
)

// 确保client类型实现了Client接口
var _ Client = (*client)(nil)

// NewClient 创建一个新的以太坊客户端
// 支持多种连接协议，会根据URL自动选择合适的传输层实现
// 参数:
//   - ctx: 用于控制客户端的生命周期的上下文对象
//   - rawURL: 连接URL，支持以下协议:
//   - http/https: 用于HTTP连接
//   - ws/wss: 用于WebSocket连接
//   - 其他: 视为IPC连接
//
// 返回:
//   - Client: 初始化好的客户端接口实现
//   - error: 可能的错误，如URL解析错误或连接建立失败
func NewClient(ctx context.Context, rawURL string) (Client, error) {
	// 解析URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse url")
	}

	var transport transport

	// 根据URL协议选择合适的传输层
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

	// 创建并返回客户端实例
	return &client{
		transport: transport,
		rawURL:    rawURL,
	}, nil
}

// NewCustomClient 使用自定义的请求者和订阅者创建客户端
// 允许用户提供自定义的请求处理和订阅处理实现
// 参数:
//   - requester: 自定义的请求处理器，实现了Requester接口
//   - subscriber: 自定义的订阅处理器，实现了Subscriber接口
//
// 返回:
//   - Client: 使用自定义处理器的客户端实例
//   - error: 创建过程中可能发生的错误
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

// transport 定义了传输层接口
// 封装了与以太坊节点通信的底层实现
// 包含以下核心功能:
//   - 请求处理: 发送JSON-RPC请求并处理响应
//   - 订阅管理: 创建和管理事件订阅
//   - 双向通信检查: 判断是否支持实时推送功能
type transport interface {
	Requester  // 继承请求处理接口
	Subscriber // 继承订阅管理接口

	// IsBidirectional 检查传输层是否支持双向通信
	// 返回true表示支持订阅和推送功能(如WebSocket)
	// 返回false表示仅支持请求响应模式(如HTTP)
	IsBidirectional() bool
}

// client 实现了Client接口的客户端结构体
// 封装了与以太坊节点交互的所有功能
type client struct {
	transport transport // 底层传输层实例，处理实际的通信细节
	rawURL    string    // 原始连接URL，记录节点的连接地址
}

// Request 发送JSON-RPC请求并返回原始响应
// 这是一个底层方法，用于发送任意JSON-RPC请求
// 参数:
//   - ctx: 控制请求生命周期的上下文对象
//   - r: 完整的JSON-RPC请求对象
//
// 返回:
//   - *jsonrpc.RawResponse: 未经处理的JSON-RPC响应
//   - error: 请求过程中可能发生的错误
func (c *client) Request(ctx context.Context, r *jsonrpc.Request) (*jsonrpc.RawResponse, error) {
	return c.transport.Request(ctx, r)
}

// Subscribe 创建一个新的订阅
// 这是一个底层方法，用于创建任意类型的订阅
// 参数:
//   - ctx: 控制订阅生命周期的上下文对象
//   - r: 订阅请求对象，通常使用eth_subscribe方法
//
// 返回:
//   - Subscription: 订阅对象，用于接收推送的事件
//   - error: 创建订阅过程中可能发生的错误
func (c *client) Subscribe(ctx context.Context, r *jsonrpc.Request) (Subscription, error) {
	return c.transport.Subscribe(ctx, r)
}

// IsBidirectional 检查客户端是否支持双向通信
// 用于判断当前连接是否支持订阅功能
// 返回:
//   - bool: true表示支持订阅(如WebSocket连接)
//     false表示不支持订阅(如HTTP连接)
func (c *client) IsBidirectional() bool {
	return c.transport.IsBidirectional()
}

// URL 返回客户端连接的URL
// 返回:
//   - string: 创建客户端时使用的原始URL
func (c *client) URL() string {
	return c.rawURL
}

// BlockNumber 获取当前区块号
// 调用eth_blockNumber方法获取最新的区块号
// 参数:
//   - ctx: 控制请求生命周期的上下文对象
//
// 返回:
//   - uint64: 最新区块的编号
//   - error: 请求过程中可能发生的错误
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

// GetTransactionCount 获取指定地址的交易数量
// 调用eth_getTransactionCount方法获取账户的nonce值
// 参数:
//   - ctx: 控制请求生命周期的上下文对象
//   - address: 要查询的账户地址
//   - numberOrTag: 区块号或特殊标签，支持:
//   - 具体区块号
//   - latest: 最新区块
//   - pending: 待处理区块
//   - earliest: 创世区块
//
// 返回:
//   - uint64: 该地址发送的交易总数(即nonce值)
//   - error: 请求过程中可能发生的错误
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

// NetVersion 获取当前网络ID
// 调用net_version方法获取以太坊网络标识符
// 参数:
//   - ctx: 控制请求生命周期的上下文对象
//
// 返回:
//   - string: 网络标识符，如:
//   - "1": 主网
//   - "3": Ropsten测试网
//   - "4": Rinkeby测试网
//   - "5": Goerli测试网
//   - error: 请求过程中可能发生的错误
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

// ChainId 获取当前链ID
// 调用eth_chainId方法获取当前链的标识符
// 参数:
//   - ctx: 控制请求生命周期的上下文对象
//
// 返回:
//   - string: 链标识符，用于交易签名
//   - error: 请求过程中可能发生的错误
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

// BlockByNumber 通过区块号获取区块信息
// 调用eth_getBlockByNumber方法获取指定区块的详细信息
// 参数:
//   - ctx: 控制请求生命周期的上下文对象
//   - number: 要查询的区块号
//   - full: 是否返回完整的交易信息
//   - true: 返回完整的交易对象
//   - false: 只返回交易哈希
//
// 返回:
//   - *eth.Block: 区块信息，包含区块头和交易列表
//   - error: 请求过程中可能发生的错误
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

// EstimateGas 估算交易执行所需的gas量
// 调用eth_estimateGas方法模拟交易执行
// 参数:
//   - ctx: 控制请求生命周期的上下文对象
//   - msg: 待估算的交易对象
//
// 返回:
//   - uint64: 预估的gas用量
//   - error: 请求过程中可能发生的错误
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

// SendRawTransaction 发送已签名的原始交易数据到网络
// 参数:
//   - ctx: 控制请求生命周期的上下文对象
//   - msg: 已签名的交易数据的十六进制字符串
//
// 返回:
//   - string: 交易哈希
//   - error: 请求过程中可能发生的错误
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

// MaxPriorityFeePerGas 获取当前建议的最大优先费用(EIP-1559)
// 参数:
//   - ctx: 控制请求生命周期的上下文对象
//
// 返回:
//   - uint64: 建议的最大优先费用(单位: wei)
//   - error: 请求过程中可能发生的错误
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

// GasPrice 获取当前网络的gas价格
// 参数:
//   - ctx: 控制请求生命周期的上下文对象
//
// 返回:
//   - uint64: 当前gas价格(单位: wei)
//   - error: 请求过程中可能发生的错误
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

// BlockByNumberOrTag 通过区块号或标签获取区块信息
// 参数:
//   - ctx: 控制请求生命周期的上下文对象
//   - numberOrTag: 区块号或特殊标签，支持:
//   - 具体区块号
//   - latest: 最新区块
//   - pending: 待处理区块
//   - earliest: 创世区块
//   - full: 是否返回完整的交易信息
//   - true: 返回完整的交易对象
//   - false: 只返回交易哈希
//
// 返回:
//   - *eth.Block: 区块信息，包含区块头和交易列表
//   - error: 请求过程中可能发生的错误
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

// BlockByHash 通过区块哈希获取区块信息
// 参数:
//   - ctx: 控制请求生命周期的上下文对象
//   - hash: 区块哈希值
//   - full: 是否返回完整的交易信息
//   - true: 返回完整的交易对象
//   - false: 只返回交易哈希
//
// 返回:
//   - *eth.Block: 区块信息，包含区块头和交易列表
//   - error: 请求过程中可能发生的错误，如哈希无效或区块未找到
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

// parseBlockResponse 解析区块响应数据
// 参数:
//   - response: JSON-RPC响应对象
//
// 返回:
//   - *eth.Block: 解析后的区块信息
//   - error: 解析过程中可能发生的错误
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

// TransactionReceipt 获取交易收据
// 参数:
//   - ctx: 控制请求生命周期的上下文对象
//   - hash: 交易哈希值
//
// 返回:
//   - *eth.TransactionReceipt: 交易收据信息，包含交易执行结果和事件日志
//   - error: 请求过程中可能发生的错误，如交易未找到或网络错误
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
		// 交易未被识别
		return nil, errors.Errorf("receipt for transaction %s not found", hash)
	}

	receipt := eth.TransactionReceipt{}
	err = json.Unmarshal(response.Result, &receipt)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal result")
	}

	return &receipt, nil
}

// Logs 获取符合过滤条件的日志
// 参数:
//   - ctx: 控制请求生命周期的上下文对象
//   - filter: 日志过滤器，用于指定日志的查询条件
//     包括区块范围、合约地址和事件主题等
//
// 返回:
//   - []eth.Log: 符合过滤条件的日志列表
//   - error: 请求过程中可能发生的错误，如参数无效或网络错误
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

// TransactionByHash 通过交易哈希获取交易信息
// 参数:
//   - ctx: 控制请求生命周期的上下文对象
//   - hash: 交易哈希值，用于唯一标识一个交易
//
// 返回:
//   - *eth.Transaction: 交易详细信息，包含发送方、接收方、金额等
//   - error: 请求过程中可能发生的错误，如哈希无效或交易未找到
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
		// 交易未被识别
		return nil, ErrTransactionNotFound
	}

	tx := eth.Transaction{}
	err = tx.UnmarshalJSON(response.Result)
	return &tx, err
}

// SubscribeNewHeads 订阅新区块头事件
// 参数:
//   - ctx: 控制订阅生命周期的上下文对象
//
// 返回:
//   - Subscription: 订阅对象，用于接收新区块头通知
//   - error: 创建订阅过程中可能发生的错误
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

// SubscribeNewPendingTransactions 订阅新的待处理交易事件
// 参数:
//   - ctx: 控制订阅生命周期的上下文对象
//
// 返回:
//   - Subscription: 订阅对象，用于接收新交易通知
//   - error: 创建订阅过程中可能发生的错误
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

// applyContext 将上下文信息应用到请求中
// 参数:
//   - ctx: 上下文对象
//   - request: 要修改的请求对象
//
// 功能:
//   - 从上下文中提取请求ID并应用到请求对象中
func applyContext(ctx context.Context, request *jsonrpc.Request) {
	if id := requestIDFromContext(ctx); id != nil {
		request.ID = *id
	}
}

// GetBalance 获取指定地址账户的余额
// 调用eth_getBalance方法获取账户在指定区块的余额
// 参数:
//   - ctx: 控制请求生命周期的上下文对象
//   - address: 要查询的账户地址
//   - numberOrTag: 区块号或标签（如"latest", "earliest", "pending"）
//
// 返回:
//   - uint64: 以wei为单位的账户余额
//   - error: 请求过程中可能发生的错误
func (c *client) GetBalance(ctx context.Context, address eth.Address, numberOrTag eth.BlockNumberOrTag) (uint64, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getBalance",
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

	return q.UInt64(), nil
}
