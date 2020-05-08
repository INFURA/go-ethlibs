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
	request     *jsonrpc.Request
	response    *jsonrpc.RawResponse
	chResult    chan *jsonrpc.RawResponse
	chError     chan error
	chAbandoned chan struct{}
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

	readMu      sync.Mutex
	readMessage readMessageFunc

	writeMu      sync.Mutex
	writeMessage writeMessageFunc
}

func (t *loopingTransport) loop() {
	g, ctx := errgroup.WithContext(t.ctx)

	// Reader
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

			// is it a request, notification, or response?
			msg, err := jsonrpc.Unmarshal(payload)
			if err != nil {
				return errors.Wrap(err, "unrecognized message from backend websocket connection")
			}

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

	// Processor
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
