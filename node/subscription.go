package node

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/INFURA/go-ethlibs/jsonrpc"
)

// subscription 表示一个以太坊订阅实例
type subscription struct {
	response        *jsonrpc.RawResponse    // 订阅请求的原始响应
	subscriptionID  string                  // 订阅的唯一标识符
	notificationsCh chan *jsonrpc.Notification // 用于向外部发送通知的通道
	dispatchCh      chan *jsonrpc.Notification // 用于内部分发通知的通道
	signalCh        chan struct{}            // 用于发送停止信号的通道
	stoppedCh       chan struct{}            // 用于指示订阅已停止的通道
	conn            Requester                // 用于发送请求的连接接口
}

// Response 返回订阅请求的原始响应
func (s *subscription) Response() *jsonrpc.RawResponse {
	return s.response
}

// ID 返回订阅的唯一标识符
func (s *subscription) ID() string {
	return s.subscriptionID
}

// Ch 返回用于接收通知的只读通道
func (s *subscription) Ch() <-chan *jsonrpc.Notification {
	return s.notificationsCh
}

// SubscriptionParams 定义了订阅通知的参数结构
type SubscriptionParams struct {
	Subscription string          `json:"subscription"` // 订阅ID
	Result       json.RawMessage `json:"result"`      // 订阅结果的原始JSON数据
}

// Unsubscribe 取消当前订阅
// ctx: 用于控制请求超时的上下文
// 返回: 如果取消订阅成功则返回nil，否则返回错误
func (s *subscription) Unsubscribe(ctx context.Context) error {
	request := jsonrpc.Request{
		ID: jsonrpc.ID{
			Str: s.subscriptionID,
		},
		Method: "eth_unsubscribe",
		Params: jsonrpc.MustParams(s.subscriptionID),
	}

	response, err := s.conn.Request(ctx, &request)
	if err != nil {
		return errors.Wrap(err, "unsubscribe failed")
	}

	if response.Error != nil {
		return errors.Errorf("%s", string(*response.Error))
	}

	return nil
}

// newSubscription 创建一个新的订阅实例
// response: 订阅请求的原始响应
// id: 订阅的唯一标识符
// r: 用于发送请求的连接接口
// 返回: 新创建的订阅实例
func newSubscription(response *jsonrpc.RawResponse, id string, r Requester) *subscription {
	s := subscription{
		response:        response,
		subscriptionID:  id,
		notificationsCh: make(chan *jsonrpc.Notification),
		dispatchCh:      make(chan *jsonrpc.Notification),
		signalCh:        make(chan struct{}),
		stoppedCh:       make(chan struct{}),
		conn:            r,
	}

	// 启动一个goroutine来将通知从dispatchCh移动到notificationsCh
	// 我们需要这个中间通道和两个信号通道来确保所有内容都按正确的顺序关闭
	// 并且没有人试图写入已关闭的通道。参考 https://go101.org/article/channel-closing.html 的变体#5
	go func() {
		defer func() {
			// 关闭stopped通道，向所有人发出信号表示这个goroutine不再运行
			// 这将导致.dispatch和.stop在不阻塞的情况下返回
			close(s.stoppedCh)

			// 然后关闭notifications通道，以便subscription.Ch()的所有消费者都被解除阻塞
			close(s.notificationsCh)
		}()

		for {
			select {
			case <-s.signalCh:
				// 收到停止信号，可以结束这个goroutine
				return
			case n := <-s.dispatchCh:
				// 有通知需要分发到外部通道
				select {
				case s.notificationsCh <- n:
					// 成功分发通知
				case <-s.signalCh:
					// 在尝试写入通知时收到停止信号
					// 可以结束这个goroutine
					return
				}
			}
		}
	}()

	return &s
}

// dispatch 将通知分发到通知通道
// ctx: 用于控制分发操作的上下文
// n: 要分发的通知
func (s *subscription) dispatch(ctx context.Context, n jsonrpc.Notification) {
	// 我们收到了一个需要分发到notificationsCh的通知
	// 但是我们不能直接在这里进行分发，因为我们只希望一个goroutine在这个通道上发送通知
	// 这样可以确保通道可以被干净地停止
	// 相反，我们将它写入一个中间工作/分发通道
	//
	// 重要说明：这个函数的`n`参数故意是整个结构体的值拷贝而不是指针
	// 这是为了确保这个函数拥有写入分发通道的指针的所有权

	select {
	case <-ctx.Done():
		// 调用.dispatch的父上下文已结束，我们可以放弃分发这个通知
		return
	case s.dispatchCh <- &n:
		// 通知现在在内部分发通道中，将由上面newSubscription中创建的goroutine处理
		return
	case <-s.stoppedCh:
		// 这个订阅已经被停止，所以我们可以放弃对它的任何写入
		return
	}
}

// stop 停止订阅的通知分发
// ctx: 用于控制停止操作的上下文
func (s *subscription) stop(ctx context.Context) {
	select {
	case <-ctx.Done():
		// 调用上下文已结束，可能是因为客户端正在关闭
		return
	case <-s.stoppedCh:
		// 我们已经被停止了，可能是由.stop()的另一个调用者触发的
		return
	case s.signalCh <- struct{}{}:
		// 我们成功地向newSubscription的goroutine发送了停止信号
		return
	}
}
