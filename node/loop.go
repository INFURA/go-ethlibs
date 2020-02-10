package node

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/INFURA/go-ethlibs/jsonrpc"
)

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

type connCloser interface {
	// Close closes the connection.
	// Any blocked Read or Write operations will be unblocked and return errors.
	Close() error

	// SetReadDeadline sets the deadline for future Read calls
	// and any currently-blocked Read call.
	// A zero value for t means Read will not time out.
	SetReadDeadline(t time.Time) error

	// SetWriteDeadline sets the deadline for future Write calls
	// and any currently-blocked Write call.
	// Even if write times out, it may return n > 0, indicating that
	// some of the data was successfully written.
	// A zero value for t means Write will not time out.
	SetWriteDeadline(t time.Time) error
}

type readMessageFunc func() ([]byte, error)
type writeMessageFunc func(payload []byte) error

type subscriptionRequest struct {
	request  *jsonrpc.Request
	response *jsonrpc.Response
	chResult chan *subscription
	chError  chan error
}

type outboundRequest struct {
	request  *jsonrpc.Request
	response *jsonrpc.RawResponse
	chResult chan *jsonrpc.RawResponse
	chError  chan error
}

type loopingTransport struct {
	conn connCloser
	ctx  context.Context

	counter                uint64
	chToBackend            chan jsonrpc.Request
	chSubscriptionRequests chan *subscriptionRequest
	chOutboundRequests     chan *outboundRequest

	subscriptonRequests map[jsonrpc.ID]*subscriptionRequest
	outboundRequests    map[jsonrpc.ID]*outboundRequest
	requestMu           sync.RWMutex

	subscriptions   map[string]*subscription
	subscriptionsMu sync.RWMutex

	readMessage  readMessageFunc
	writeMessage writeMessageFunc
}

func (t *loopingTransport) loop() {
	g, ctx := errgroup.WithContext(t.ctx)

	// Reader
	g.Go(func() error {
		for {
			payload, err := t.readMessage()
			if err != nil {
				if ctx.Err() == context.Canceled {
					log.Printf("[DEBUG] Context cancelled during read")
					return nil
				}

				return errors.Wrap(err, "error reading message")
			}

			if payload == nil {
				continue
			}
			// log.Printf("[SPAM] read: %s", string(payload))

			// is it a request, notification, or response?
			msg, err := jsonrpc.Unmarshal(payload)
			if err != nil {
				return errors.Wrap(err, "unrecognized message from backend websocket connection")
			}

			switch msg := msg.(type) {
			case *jsonrpc.RawResponse:
				// log.Printf("[SPAM] response: %v", msg)

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
						subCtx, subCancel := context.WithCancel(t.ctx)

						subscription := &subscription{
							response:       &patchedResponse,
							subscriptionID: result,
							ch:             make(chan *jsonrpc.Notification),
							conn:           t,
							ctx:            subCtx,
							cancel:         subCancel,
						}
						t.subscriptionsMu.Lock()
						t.subscriptions[result] = subscription
						t.subscriptionsMu.Unlock()

						go func() {
							select {
							case <-ctx.Done():
								return
							case start.chResult <- subscription:
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

					go func(r *jsonrpc.RawResponse) {
						patchedResponse := *r
						patchedResponse.ID = outbound.request.ID
						select {
						case <-ctx.Done():
							return
						case outbound.chResult <- &patchedResponse:
							return
						}

					}(msg)
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

				t.subscriptionsMu.RLock()
				if subscription, ok := t.subscriptions[sp.Subscription]; ok {
					go func(n *jsonrpc.Notification) {
						_copy := *n
						select {
						case <-ctx.Done():
							return
						case subscription.ch <- &_copy:
							return
						}

					}(msg)
				}
				t.subscriptionsMu.RUnlock()
			}
		}
	})

	// Writer
	g.Go(func() error {
		for {
			select {
			case r := <-t.chToBackend:
				// log.Printf("[SPAM] Writing %v", r)
				b, err := json.Marshal(&r)
				if err != nil {
					return errors.Wrap(err, "error marshalling request for backend")
				}

				// log.Printf("[SPAM] Writing %s", string(b))
				err = t.writeMessage(b)
				if err != nil {
					if ctx.Err() == context.Canceled {
						log.Printf("[DEBUG] Context cancelled during write")
						return nil
					}

					return errors.Wrap(err, "error writing to backend websocket connection")
				}

			case <-ctx.Done():
				return nil
			}
		}
	})

	// Processor
	g.Go(func() error {
		for {
			select {
			// subscriptions
			case s := <-t.chSubscriptionRequests:
				// log.Printf("[SPAM] Sub request %v", s)
				id := t.nextID()
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
				id := t.nextID()
				proxy := *o.request
				proxy.ID = id
				// log.Printf("[SPAM] outbound proxied request %v", proxy)

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

	// Aborter
	g.Go(func() error {
		select {
		case <-ctx.Done():
			log.Printf("[DEBUG] Context done, setting deadlines to now")
			_ = t.conn.SetReadDeadline(time.Now())
			_ = t.conn.SetWriteDeadline(time.Now())
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
		sub.mu.Lock()
		sub.err = err
		sub.mu.Unlock()
		sub.cancel()
		delete(t.subscriptions, id)
	}
	t.subscriptionsMu.Unlock()

	_ = t.conn.Close()
}

func (t *loopingTransport) nextID() jsonrpc.ID {
	return jsonrpc.ID{
		Num: atomic.AddUint64(&t.counter, 1),
	}
}

func (t *loopingTransport) Request(ctx context.Context, r *jsonrpc.Request) (*jsonrpc.RawResponse, error) {

	outbound := outboundRequest{
		request:  r,
		chResult: make(chan *jsonrpc.RawResponse),
		chError:  make(chan error),
	}

	defer func() {
		close(outbound.chResult)
		close(outbound.chError)
	}()

	select {
	case t.chOutboundRequests <- &outbound:
		// log.Printf("[SPAM] outbound request sent")
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "context finished waiting for response")
	}

	select {
	case response := <-outbound.chResult:
		return response, nil
	case err := <-outbound.chError:
		return nil, err
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "context finished waiting for response")
	}
}

func (t *loopingTransport) Subscribe(ctx context.Context, r *jsonrpc.Request) (Subscription, error) {
	if r.Method != "eth_subscribe" && r.Method != "parity_subscribe" {
		return nil, errors.New("request is not a subscription request")
	}

	start := subscriptionRequest{
		request:  r,
		chResult: make(chan *subscription),
		chError:  make(chan error),
	}

	defer close(start.chResult)
	defer close(start.chError)

	select {
	case t.chSubscriptionRequests <- &start:
		// log.Printf("[SPAM] start request sent")
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "context finished waiting for subscription")
	}

	select {
	case s := <-start.chResult:
		return s, nil
	case err := <-start.chError:
		return nil, err
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "context finished waiting for subscription")
	}
}

func (t *loopingTransport) IsBidirectional() bool {
	return true
}
