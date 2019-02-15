package node

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/INFURA/eth/pkg/eth"
	"github.com/INFURA/eth/pkg/jsonrpc"
)

// Backend represents a backend websocket connection to an ethereum client node.
type Backend interface {
	// URL returns the backends URL
	URL() string

	// BlockNumber returns the current block number at head
	BlockNumber(ctx context.Context) (uint64, error)

	// BlockByNumber can be used to get a block by it's number
	BlockByNumber(ctx context.Context, number uint64, full bool) (*eth.Block, error)

	// BlockByHash can be used to get a block by it's hash
	BlockByHash(ctx context.Context, hash string, full bool) (*eth.Block, error)

	// Request method can be used by downstream consumers of ChangeEvent to make generic JSONRPC requests
	Request(ctx context.Context, r *jsonrpc.Request) (*jsonrpc.RawResponse, error)

	// NewHeads should generally not be used externally
	NewHeads(ctx context.Context) (Subscription, error)
}

type backend struct {
	url  string
	conn *websocket.Conn
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
}

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

// NewBackend creates a Backend from the passed in URL.  Use the supplied Context to shutdown the connection by
// cancelling or otherwise aborting the context.
func NewBackend(ctx context.Context, url string) (Backend, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	b := backend{
		conn:                   conn,
		url:                    url,
		ctx:                    ctx,
		counter:                rand.Uint64(),
		chToBackend:            make(chan jsonrpc.Request),
		chSubscriptionRequests: make(chan *subscriptionRequest),
		chOutboundRequests:     make(chan *outboundRequest),
		subscriptonRequests:    make(map[jsonrpc.ID]*subscriptionRequest),
		outboundRequests:       make(map[jsonrpc.ID]*outboundRequest),
		subscriptions:          make(map[string]*subscription),
	}

	go b.loop()
	return &b, nil
}

func (b *backend) loop() {
	g, ctx := errgroup.WithContext(b.ctx)

	// Reader
	g.Go(func() error {
		for {
			typ, r, err := b.conn.NextReader()
			if err != nil {
				if ctx.Err() == context.Canceled {
					log.Printf("[DEBUG] Context cancelled during read")
					return nil
				}
				return errors.Wrap(err, "error reading from backend")
			}

			if typ != websocket.TextMessage {
				continue
			}

			payload, err := ioutil.ReadAll(r)
			if err != nil {
				if ctx.Err() == context.Canceled {
					log.Printf("[DEBUG] Context cancelled during read")
					return nil
				}

				return errors.Wrap(err, "error reading from backend")
			}

			// log.Printf("[SPAM] read: %s", string(payload))

			// is it a request, notification, or response?
			msg, err := jsonrpc.Unmarshal(payload)
			if err != nil {
				return errors.Wrap(err, "unrecognized message from backend")
			}

			switch msg := msg.(type) {
			case *jsonrpc.RawResponse:
				// log.Printf("[SPAM] response: %v", msg)

				// subscriptions
				b.requestMu.Lock()
				if start, ok := b.subscriptonRequests[msg.ID]; ok {
					delete(b.subscriptonRequests, msg.ID)
					b.requestMu.Unlock()

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
						return errors.Wrap(err, "unparsable result from backend")
					}

					// log.Printf("[SPAM]: Result: %v", result)

					switch result := result.(type) {
					case string:
						subCtx, subCancel := context.WithCancel(b.ctx)

						subscription := &subscription{
							response:       &patchedResponse,
							subscriptionID: result,
							ch:             make(chan *jsonrpc.Notification),
							backend:        b,
							ctx:            subCtx,
							cancel:         subCancel,
						}
						b.subscriptionsMu.Lock()
						b.subscriptions[result] = subscription
						b.subscriptionsMu.Unlock()

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
				if outbound, ok := b.outboundRequests[msg.ID]; ok {
					delete(b.outboundRequests, msg.ID)
					b.requestMu.Unlock()

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
				b.requestMu.Unlock()

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

				b.subscriptionsMu.RLock()
				if subscription, ok := b.subscriptions[sp.Subscription]; ok {
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
				b.subscriptionsMu.RUnlock()
			}
		}

	})

	// Writer
	g.Go(func() error {
		for {
			select {
			case r := <-b.chToBackend:
				// log.Printf("[SPAM] Writing %v", r)
				err := b.conn.WriteJSON(&r)
				if err != nil {
					if ctx.Err() == context.Canceled {
						log.Printf("[DEBUG] Context cancelled during write")
						return nil
					}

					return errors.Wrap(err, "error writing to backend")
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
			case s := <-b.chSubscriptionRequests:
				// log.Printf("[SPAM] Sub request %v", s)
				id := b.nextID()
				proxy := *s.request
				proxy.ID = id

				b.requestMu.Lock()
				b.subscriptonRequests[id] = s
				b.requestMu.Unlock()

				select {
				case <-ctx.Done():
					return ctx.Err()
				case b.chToBackend <- proxy:
					continue
				}

			// outbound requests
			case o := <-b.chOutboundRequests:
				// log.Printf("[SPAM] outbound request %v", *o)
				id := b.nextID()
				proxy := *o.request
				proxy.ID = id
				// log.Printf("[SPAM] outbound proxied request %v", proxy)

				b.requestMu.Lock()
				b.outboundRequests[id] = o
				b.requestMu.Unlock()

				select {
				case <-ctx.Done():
					return ctx.Err()
				case b.chToBackend <- proxy:
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
			_ = b.conn.SetReadDeadline(time.Now())
			_ = b.conn.SetWriteDeadline(time.Now())
		}

		return nil
	})

	err := g.Wait()
	if err == nil {
		err = context.Canceled
	}

	// let's clean up all the remaining subscriptions
	b.subscriptionsMu.Lock()
	for id, sub := range b.subscriptions {
		sub.mu.Lock()
		sub.err = err
		sub.mu.Unlock()
		sub.cancel()
		delete(b.subscriptions, id)
	}
	b.subscriptionsMu.Unlock()
}

func (b *backend) nextID() jsonrpc.ID {
	return jsonrpc.ID{
		Num: atomic.AddUint64(&b.counter, 1),
	}
}

func (b *backend) URL() string {
	return b.url
}

func (b *backend) Request(ctx context.Context, r *jsonrpc.Request) (*jsonrpc.RawResponse, error) {

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
	case b.chOutboundRequests <- &outbound:
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

func (b *backend) Subscribe(ctx context.Context, r *jsonrpc.Request) (Subscription, error) {
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
	case b.chSubscriptionRequests <- &start:
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

func (b *backend) BlockNumber(ctx context.Context) (uint64, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_blockNumber",
	}

	response, err := b.Request(ctx, &request)
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

func (b *backend) BlockByNumber(ctx context.Context, number uint64, full bool) (*eth.Block, error) {
	n := eth.QuantityFromUInt64(number)

	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getBlockByNumber",
		Params: jsonrpc.MustParams(&n, full),
	}

	// log.Printf("[SPAM] params: [%s, %s]", string(request.Params[0]), string(request.Params[1]))

	response, err := b.Request(ctx, &request)
	if err != nil {
		return nil, errors.Wrap(err, "could not make request")
	}

	return b.parseBlockResponse(response)
}

func (b *backend) parseBlockResponse(response *jsonrpc.RawResponse) (*eth.Block, error) {
	if response.Error != nil {
		return nil, errors.New(string(*response.Error))
	}

	// log.Printf("[SPAM] Result: %s", string(response.Result))

	block := eth.Block{}
	err := json.Unmarshal(response.Result, &block)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode block result")
	}

	return &block, nil
}

func (b *backend) BlockByHash(ctx context.Context, hash string, full bool) (*eth.Block, error) {
	// TODO: Support full=false, requires handling block.transactions as just strings instead of objects
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getBlockByHash",
		Params: jsonrpc.MustParams(hash, full),
	}

	response, err := b.Request(ctx, &request)
	if err != nil {
		return nil, errors.Wrap(err, "could not make request")
	}

	return b.parseBlockResponse(response)
}

func (b *backend) NewHeads(ctx context.Context) (Subscription, error) {
	r := jsonrpc.Request{
		JSONRPC: "2.0",
		ID:      jsonrpc.ID{Str: "test", IsString: true},
		Method:  "eth_subscribe",
		Params:  jsonrpc.MustParams("newHeads"),
	}

	return b.Subscribe(ctx, &r)
}
