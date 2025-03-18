// Package node 提供以太坊节点连接和通信功能
package node

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/INFURA/go-ethlibs/jsonrpc"
)

// newLoopingTransport 创建一个新的循环传输层实例
// ctx 用于控制传输层的生命周期
// conn 实现了connCloser接口的连接对象
// readMessage 用于读取消息的函数
// writeMessage 用于写入消息的函数
// 返回初始化好的循环传输层实例
func newLoopingTransport(ctx context.Context, conn connCloser, readMessage readMessageFunc, writeMessage writeMessageFunc) *loopingTransport {
	t := loopingTransport{
		conn:                   conn,
		ctx:                    ctx,
		counter:                rand.Uint64(),
		chToBackend:            make(chan jsonrpc.Request),
		chSubscriptionRequests: make(chan *subscriptionRequest),
		chOutboundRequests:     make(chan *outboundRequest),
		subscriptonRequests:    make(map[jsonrpc.ID]*subscriptionRequest),
		outboundRequests:       make(map[jsonrpc.ID]*outboundRequest),
		subscriptions:          make(map[string]*subscription),
		readMessage:            readMessage,
		writeMessage:           writeMessage,
	}

	go t.loop()
	return &t
}

// connCloser 定义了连接关闭和超时设置的接口
type connCloser interface {
	// Close 关闭连接
	// 任何被阻塞的读写操作都将被解除阻塞并返回错误
	Close() error

	// SetReadDeadline 设置未来读取操作的截止时间
	// 对当前被阻塞的读取操作也生效
	// t为零值表示读取操作不会超时
	SetReadDeadline(t time.Time) error

	// SetWriteDeadline 设置未来写入操作的截止时间
	// 对当前被阻塞的写入操作也生效
	// 即使写入超时，也可能返回n>0，表示部分数据已成功写入
	// t为零值表示写入操作不会超时
	SetWriteDeadline(t time.Time) error
}

// readMessageFunc 定义了读取消息的函数类型
type readMessageFunc func() ([]byte, error)

// writeMessageFunc 定义了写入消息的函数类型
type writeMessageFunc func(payload []byte) error

// subscriptionRequest 表示订阅请求的结构体
type subscriptionRequest struct {
	request  *jsonrpc.Request   // JSON-RPC请求对象
	response *jsonrpc.Response  // JSON-RPC响应对象
	chResult chan *subscription // 用于接收订阅结果的通道
	chError  chan error         // 用于接收错误的通道
}

// outboundRequest 表示出站请求的结构体
type outboundRequest struct {
	request     *jsonrpc.Request          // JSON-RPC请求对象
	response    *jsonrpc.RawResponse      // 原始JSON-RPC响应对象
	chResult    chan *jsonrpc.RawResponse // 用于接收响应结果的通道
	chError     chan error                // 用于接收错误的通道
	chAbandoned chan struct{}             // 用于标记请求是否被放弃的通道
}

// loopingTransport 实现了循环传输层的核心功能
type loopingTransport struct {
	conn connCloser      // 底层连接对象
	ctx  context.Context // 控制传输层生命周期的上下文

	counter                uint64                    // 请求计数器
	chToBackend            chan jsonrpc.Request      // 发送到后端的请求通道
	chSubscriptionRequests chan *subscriptionRequest // 订阅请求通道
	chOutboundRequests     chan *outboundRequest     // 出站请求通道

	subscriptonRequests map[jsonrpc.ID]*subscriptionRequest // 订阅请求映射表
	outboundRequests    map[jsonrpc.ID]*outboundRequest     // 出站请求映射表
	requestMu           sync.RWMutex                        // 请求映射表的读写锁

	subscriptions   map[string]*subscription // 活跃订阅映射表
	subscriptionsMu sync.RWMutex             // 订阅映射表的读写锁

	readMu      sync.Mutex      // 读取操作的互斥锁
	readMessage readMessageFunc // 读取消息的函数

	writeMu      sync.Mutex       // 写入操作的互斥锁
	writeMessage writeMessageFunc // 写入消息的函数
}

// loop 启动循环传输层的主循环，处理消息的读取、写入和订阅管理
func (t *loopingTransport) loop() {
	g, ctx := errgroup.WithContext(t.ctx)

	// 启动读取器协程
	g.Go(func() error {
		for {
			t.readMu.Lock()
			payload, err := t.readMessage()
			t.readMu.Unlock()
			if err != nil {
				if ctx.Err() == context.Canceled {
					return nil
				}

				return errors.Wrap(err, "error reading message")
			}

			if payload == nil {
				continue
			}
			// log.Printf("[SPAM] read: %s", string(payload))

			// 解析消息类型（请求、通知或响应）
			msg, err := jsonrpc.Unmarshal(payload)
			if err != nil {
				return errors.Wrap(err, "unrecognized message from backend websocket connection")
			}

			// 根据消息类型进行不同的处理
			switch msg := msg.(type) {
			case *jsonrpc.RawResponse:
				// log.Printf("[SPAM] response: %p", msg)

				// subscriptions
				t.requestMu.Lock()
				if start, ok := t.subscriptonRequests[msg.ID]; ok {
					delete(t.subscriptonRequests, msg.ID)
					t.requestMu.Unlock()

					patchedResponse := *msg
					patchedResponse.ID = start.request.ID

					if patchedResponse.Result == nil || patchedResponse.Error != nil {
						select {
						case <-ctx.Done():
							continue
						case start.chError <- errors.New("Error w/ subscription"):
							continue
						}
					}

					var result interface{}
					err = json.Unmarshal(patchedResponse.Result, &result)
					if err != nil {
						return errors.Wrap(err, "unparsable result from backend websocket connection")
					}

					// log.Printf("[SPAM]: Result: %v", result)

					switch result := result.(type) {
					case string:
						sub := newSubscription(&patchedResponse, result, t)
						t.subscriptionsMu.Lock()
						t.subscriptions[result] = sub
						t.subscriptionsMu.Unlock()

						go func() {
							select {
							case <-ctx.Done():
								return
							case start.chResult <- sub:
								return
							}
						}()
						continue
					default:
						select {
						case <-ctx.Done():
							continue
						case start.chError <- errors.New("Non-string subscription id"):
							continue
						}
					}
				}

				// other responses
				if outbound, ok := t.outboundRequests[msg.ID]; ok {
					delete(t.outboundRequests, msg.ID)
					t.requestMu.Unlock()

					go func(o *outboundRequest, r *jsonrpc.RawResponse) {
						patchedResponse := *r
						patchedResponse.ID = o.request.ID
						select {
						case <-ctx.Done():
							return
						case <-o.chAbandoned:
							// request was abandoned (e.g. client disconnected)
							log.Printf("[WARN] request abandoned %v %v", r.ID, o.request.ID)
							return
						case o.chResult <- &patchedResponse:
							return
						}
					}(outbound, msg)
					continue
				}
				t.requestMu.Unlock()

			case *jsonrpc.Request:
				// log.Printf("[SPAM] request: %v", msg)
			case *jsonrpc.Notification:
				// log.Printf("[SPAM] notif: %v", msg)
				if msg.Method != "eth_subscription" {
					continue
				}

				sp := SubscriptionParams{}
				err := json.Unmarshal(msg.Params, &sp)
				if err != nil {
					log.Printf("[WARN] eth_subscription Notification not decoded: %v", err)
					continue
				}

				go func(n jsonrpc.Notification) {
					t.subscriptionsMu.RLock()
					defer t.subscriptionsMu.RUnlock()
					if subscription, ok := t.subscriptions[sp.Subscription]; ok {
						subscription.dispatch(ctx, n)
					}
				}(*msg)
			}
		}
	})

	// 启动写入器协程
	g.Go(func() error {
		for {
			select {
			case r := <-t.chToBackend:
				// log.Printf("[SPAM] Writing %v", r)
				b, err := json.Marshal(&r)
				if err != nil {
					return errors.Wrap(err, "error marshalling request for backend")
				}

				if r.Method == "eth_unsubscribe" {
					// process the unsubscribe here just in case someone
					// manually creates the eth_unsubscribe RPC versus calling subscription.Unsubscribe()
					var id string
					if r.Params.UnmarshalInto(&id) == nil {
						log.Printf("[DEBUG] removing subscription id %s", id)
						t.subscriptionsMu.Lock()
						if sub, ok := t.subscriptions[id]; ok {
							sub.stop(ctx)
							delete(t.subscriptions, id)
						}
						t.subscriptionsMu.Unlock()
					}
				}

				// log.Printf("[SPAM] Writing %s", string(b))
				t.writeMu.Lock()
				err = t.writeMessage(b)
				t.writeMu.Unlock()
				if err != nil {
					if ctx.Err() == context.Canceled {
						return nil
					}

					return errors.Wrap(err, "error writing to backend websocket connection")
				}

			case <-ctx.Done():
				return nil
			}
		}
	})

	// 启动处理器协程，处理订阅请求和出站请求
	g.Go(func() error {
		for {
			select {
			// subscriptions
			case s := <-t.chSubscriptionRequests:
				// log.Printf("[SPAM] Sub request %v", s)
				id := t.nextID(s.request.ID)
				proxy := *s.request
				proxy.ID = id

				t.requestMu.Lock()
				t.subscriptonRequests[id] = s
				t.requestMu.Unlock()

				select {
				case <-ctx.Done():
					return ctx.Err()
				case t.chToBackend <- proxy:
					continue
				}

			// outbound requests
			case o := <-t.chOutboundRequests:
				// log.Printf("[SPAM] outbound request %v", *o)
				id := t.nextID(o.request.ID)
				proxy := *o.request
				proxy.ID = id
				// log.Printf("[DEBUG] outbound proxied request method %s ID %v was %v", proxy.Method, proxy.ID, o.request.ID)

				t.requestMu.Lock()
				t.outboundRequests[id] = o
				t.requestMu.Unlock()

				select {
				case <-ctx.Done():
					return ctx.Err()
				case t.chToBackend <- proxy:
					continue
				}

			case <-ctx.Done():
				return nil
			}
		}
	})

	// Closer
	g.Go(func() error {
		select {
		case <-ctx.Done():
			_ = t.conn.Close()
		}

		return nil
	})

	err := g.Wait()
	if err == nil {
		err = context.Canceled
	}

	// let's clean up all the remaining subscriptions
	t.subscriptionsMu.Lock()
	for id, sub := range t.subscriptions {
		// don't pass in our ctx here, it's already been stopped
		sub.stop(context.Background())
		delete(t.subscriptions, id)
	}
	t.subscriptionsMu.Unlock()

	_ = t.conn.Close()
}

func (t *loopingTransport) nextID(seed jsonrpc.ID) jsonrpc.ID {
	n := atomic.AddUint64(&t.counter, 1)
	if seed.IsString {
		return jsonrpc.ID{
			Num:      0,
			Str:      fmt.Sprintf("%s-%d", seed.Str, n),
			IsString: true,
		}
	}
	return jsonrpc.ID{
		Num: n,
	}
}

func (t *loopingTransport) Request(ctx context.Context, r *jsonrpc.Request) (*jsonrpc.RawResponse, error) {
	select {
	case <-t.ctx.Done():
		return nil, errors.Wrap(t.ctx.Err(), "transport context finished")
	default:
		// transport context is still valid, we can process this request
	}

	owned, err := copyRequest(r)
	if err != nil {
		return nil, err
	}
	outbound := &outboundRequest{
		request:     &owned,
		chResult:    make(chan *jsonrpc.RawResponse),
		chError:     make(chan error),
		chAbandoned: make(chan struct{}),
	}

	defer func() {
		close(outbound.chAbandoned)
	}()

	select {
	case t.chOutboundRequests <- outbound:
		// log.Printf("[SPAM] outbound request sent")
	case <-t.ctx.Done():
		return nil, errors.Wrap(t.ctx.Err(), "transport context finished waiting for response")
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "context finished waiting for response")
	}

	select {
	case response := <-outbound.chResult:
		return response, nil
	case err := <-outbound.chError:
		return nil, err
	case <-t.ctx.Done():
		return nil, errors.Wrap(t.ctx.Err(), "transport context finished waiting for response")
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "context finished waiting for response")
	}
}

func copyRequest(request *jsonrpc.Request) (jsonrpc.Request, error) {
	copied := jsonrpc.Request{}
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(request); err != nil {
		return copied, errors.Wrap(err, "could not copy request: encoding failed")
	}
	if err := json.NewDecoder(buf).Decode(&copied); err != nil {
		return copied, errors.Wrap(err, "could not copy request: decoding failed")
	}
	return copied, nil
}

func (t *loopingTransport) Subscribe(ctx context.Context, r *jsonrpc.Request) (Subscription, error) {
	if r.Method != "eth_subscribe" && r.Method != "parity_subscribe" {
		return nil, errors.New("request is not a subscription request")
	}

	select {
	case <-t.ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "transport context finished")
	default:
		// transport context is still valid, we can process this request
	}

	owned, err := copyRequest(r)
	if err != nil {
		return nil, err
	}

	start := subscriptionRequest{
		request:  &owned,
		chResult: make(chan *subscription),
		chError:  make(chan error),
	}

	defer close(start.chResult)
	defer close(start.chError)

	select {
	case t.chSubscriptionRequests <- &start:
		// log.Printf("[SPAM] start request sent")
	case <-t.ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "transport context finished waiting for subscription")
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "context finished waiting for subscription")
	}

	select {
	case s := <-start.chResult:
		return s, nil
	case err := <-start.chError:
		return nil, err
	case <-t.ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "transport context finished waiting for subscription")
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "context finished waiting for subscription")
	}
}

func (t *loopingTransport) IsBidirectional() bool {
	return true
}
