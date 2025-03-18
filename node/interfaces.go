package node

//go:generate mockgen -source=interfaces.go -destination=mocks/node.go -package=mock
//go:generate mockgen -source=interfaces.go -destination=interfaces_mock.go -package=node

import (
	"context"

	"github.com/INFURA/go-ethlibs/eth"
	"github.com/INFURA/go-ethlibs/jsonrpc"
)

// Requester 定义了发送JSON-RPC请求的接口
type Requester interface {
	// Request 发送JSON-RPC请求并接收响应
	// ctx 上下文对象，用于控制请求的生命周期
	// r JSON-RPC请求对象
	// 返回原始JSON-RPC响应和可能的错误
	Request(ctx context.Context, r *jsonrpc.Request) (*jsonrpc.RawResponse, error)
}

// Subscriber 定义了处理以太坊订阅的接口
type Subscriber interface {
	// Subscribe 通过eth_subscribe方法创建新的订阅
	// ctx 上下文对象，用于控制订阅的生命周期
	// r JSON-RPC订阅请求对象
	// 返回订阅实例和可能的错误
	Subscribe(ctx context.Context, r *jsonrpc.Request) (Subscription, error)
}

// Client 表示与以太坊节点的连接，继承了Requester和Subscriber接口
type Client interface {
	Requester
	Subscriber

	// URL 返回当前连接的后端URL地址
	URL() string

	// BlockNumber 返回当前最新区块号
	BlockNumber(ctx context.Context) (uint64, error)

	// NetVersion 返回网络版本号
	NetVersion(ctx context.Context) (string, error)

	// ChainId 返回链ID
	ChainId(ctx context.Context) (string, error)

	// EstimateGas 估算交易所需的gas量
	// msg 待估算的交易对象
	// 返回预估的gas用量和可能的错误
	EstimateGas(ctx context.Context, msg eth.Transaction) (uint64, error)

	// MaxPriorityFeePerGas (EIP1559) 返回建议的区块小费
	MaxPriorityFeePerGas(ctx context.Context) (uint64, error)

	// GasPrice (Legacy) 返回建议的gas价格
	GasPrice(ctx context.Context) (uint64, error)

	// GetTransactionCount 获取指定地址的待处理nonce值
	// address 要查询的账户地址
	// numberOrTag 区块号或标签（如"latest", "pending"）
	// 返回nonce值和可能的错误
	GetTransactionCount(ctx context.Context, address eth.Address, numberOrTag eth.BlockNumberOrTag) (uint64, error)

	// SendRawTransaction 发送已签名的原始交易
	// msg 已签名的交易数据
	// 返回交易哈希和可能的错误
	SendRawTransaction(ctx context.Context, msg string) (string, error)

	// BlockByNumber 通过区块号获取区块信息
	// number 区块号
	// full 是否返回完整的交易信息
	// 返回区块信息和可能的错误
	BlockByNumber(ctx context.Context, number uint64, full bool) (*eth.Block, error)

	// BlockByNumberOrTag 通过区块号或标签获取区块信息
	// numberOrTag 区块号或标签（如"latest"）
	// full 是否返回完整的交易信息
	// 返回区块信息和可能的错误
	BlockByNumberOrTag(ctx context.Context, numberOrTag eth.BlockNumberOrTag, full bool) (*eth.Block, error)

	// BlockByHash 通过区块哈希获取区块信息
	// hash 区块哈希
	// full 是否返回完整的交易信息
	// 返回区块信息和可能的错误
	BlockByHash(ctx context.Context, hash string, full bool) (*eth.Block, error)

	// TransactionByHash 通过交易哈希获取交易信息
	// hash 交易哈希
	// 返回交易信息和可能的错误
	TransactionByHash(ctx context.Context, hash string) (*eth.Transaction, error)

	// SubscribeNewHeads 订阅新区块头事件
	// 返回订阅实例和可能的错误
	SubscribeNewHeads(ctx context.Context) (Subscription, error)

	// SubscribeNewPendingTransactions 订阅新的待处理交易事件
	// 返回订阅实例和可能的错误
	SubscribeNewPendingTransactions(ctx context.Context) (Subscription, error)

	// TransactionReceipt 获取指定交易的收据
	// hash 交易哈希
	// 返回交易收据和可能的错误
	TransactionReceipt(ctx context.Context, hash string) (*eth.TransactionReceipt, error)

	// Logs 返回符合过滤条件的日志数组
	// filter 日志过滤条件
	// 返回日志数组和可能的错误
	Logs(ctx context.Context, filter eth.LogFilter) ([]eth.Log, error)

	// GetBalance 返回指定地址账户的余额
	// address 要查询的账户地址
	// numberOrTag 区块号或标签（如"latest", "earliest", "pending"）
	// 返回以wei为单位的余额和可能的错误
	GetBalance(ctx context.Context, address eth.Address, numberOrTag eth.BlockNumberOrTag) (uint64, error)

	// IsBidirectional 检查传输层是否支持双向通信功能（如订阅）
	// 返回true表示支持订阅功能
	IsBidirectional() bool
}

// Subscription 定义了以太坊订阅接口
type Subscription interface {
	// Response 返回订阅的原始响应
	Response() *jsonrpc.RawResponse
	// ID 返回订阅的唯一标识符
	ID() string
	// Ch 返回用于接收订阅通知的通道
	Ch() <-chan *jsonrpc.Notification
	// Unsubscribe 取消订阅
	Unsubscribe(ctx context.Context) error
}
