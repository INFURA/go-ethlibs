package node

import (
	"context"

	"github.com/pkg/errors"

	"github.com/INFURA/go-ethlibs/jsonrpc"
)

// customTransport 自定义传输层结构体，用于处理JSON-RPC请求和订阅
type customTransport struct {
	// requester 处理单次请求的接口实现
	requester Requester
	// subscriber 处理订阅请求的接口实现
	subscriber Subscriber
}

// Request 处理单次JSON-RPC请求
// ctx 上下文对象，用于控制请求的生命周期
// r 待处理的JSON-RPC请求
// 返回JSON-RPC响应和可能的错误
func (t *customTransport) Request(ctx context.Context, r *jsonrpc.Request) (*jsonrpc.RawResponse, error) {
	return t.requester.Request(ctx, r)
}

// Subscribe 处理JSON-RPC订阅请求
// ctx 上下文对象，用于控制订阅的生命周期
// r 待处理的JSON-RPC订阅请求
// 返回订阅对象和可能的错误
func (t *customTransport) Subscribe(ctx context.Context, r *jsonrpc.Request) (Subscription, error) {
	if t.subscriber == nil {
		return nil, errors.New("subscriptions not supported over this transport")
	}

	return t.subscriber.Subscribe(ctx, r)
}

// IsBidirectional 检查传输层是否支持双向通信（订阅功能）
// 返回true表示支持订阅功能，false表示仅支持单次请求
func (t *customTransport) IsBidirectional() bool {
	return t.subscriber != nil
}

// newCustomTransport 创建新的自定义传输层实例
// requester 处理单次请求的实现
// subscriber 处理订阅请求的实现（可选）
// 返回传输层实例和可能的错误
func newCustomTransport(requester Requester, subscriber Subscriber) (*customTransport, error) {
	t := customTransport{
		requester:  requester,
		subscriber: subscriber,
	}

	return &t, nil
}
