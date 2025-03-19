// Package websocket implements a websocket connection to an Ethereum node.
//
// Deprecated: This package has been superseded by the node.Client interface.
//
// This package is frozen and no new functionality will be added.
package websocket

//go:generate mockgen -source=websocket.go -destination=mocks/websocket.go -package=mock

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/justinwongcn/go-ethlibs/eth"
	"github.com/justinwongcn/go-ethlibs/jsonrpc"
)

// Connection represents a websocket connection to a backend ethereum client node.
type Connection interface {
	// URL returns the backend URL we are connected to
	URL() string

	// BlockNumber returns the current block number at head
	BlockNumber(ctx context.Context) (uint64, error)

	// BlockByNumber can be used to get a block by its number
	BlockByNumber(ctx context.Context, number uint64, full bool) (*eth.Block, error)

	// BlockByHash can be used to get a block by its hash
	BlockByHash(ctx context.Context, hash string, full bool) (*eth.Block, error)

	// eth_getTransactionByHash can be used to get transaction by its hash
	TransactionByHash(ctx context.Context, hash string) (*eth.Transaction, error)

	// Request method can be used by downstream consumers of ChangeEvent to make generic JSONRPC requests
	Request(ctx context.Context, r *jsonrpc.Request) (*jsonrpc.RawResponse, error)

	// Subscibe method can be used to make subscription requests
	Subscribe(ctx context.Context, r *jsonrpc.Request) (Subscription, error)

	// NewHeads subscription
	NewHeads(ctx context.Context) (Subscription, error)

	// NewPendingTransactions subscriptions
	NewPendingTransaction(ctx context.Context, full ...bool) (Subscription, error)

	// TransactionReceipt for a particular transaction
	TransactionReceipt(ctx context.Context, hash string) (*eth.TransactionReceipt, error)

	// GetLogs
	GetLogs(ctx context.Context, filter eth.LogFilter) ([]eth.Log, error)
}

var (
	ErrBlockNotFound       = errors.New("block not found")
	ErrTransactionNotFound = errors.New("transaction not found")
)

type connection struct {
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
	chResult chan *subscription
	chError  chan error
}

type outboundRequest struct {
	request  *jsonrpc.Request
	chResult chan *jsonrpc.RawResponse
	chError  chan error
}

// NewConnection creates a Connection to the passed in URL.  Use the supplied Context to shutdown the connection by
// cancelling or otherwise aborting the context.
func NewConnection(ctx context.Context, url string) (Connection, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	b := connection{
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

func (c *connection) loop() {
	g, ctx := errgroup.WithContext(c.ctx)

	// Reader
	g.Go(func() error {
		for {
			typ, r, err := c.conn.NextReader()
			if err != nil {
				if ctx.Err() == context.Canceled {
					log.Printf("[DEBUG] Context cancelled during read")
					return nil
				}
				return errors.Wrap(err, "error reading from backend websocket connection")
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

				return errors.Wrap(err, "error reading from backend websocket connection")
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
				c.requestMu.Lock()
				if start, ok := c.subscriptonRequests[msg.ID]; ok {
					delete(c.subscriptonRequests, msg.ID)
					c.requestMu.Unlock()

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
						subCtx, subCancel := context.WithCancel(c.ctx)

						subscription := &subscription{
							response:       &patchedResponse,
							subscriptionID: result,
							ch:             make(chan *jsonrpc.Notification),
							conn:           c,
							ctx:            subCtx,
							cancel:         subCancel,
						}
						c.subscriptionsMu.Lock()
						c.subscriptions[result] = subscription
						c.subscriptionsMu.Unlock()

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
				if outbound, ok := c.outboundRequests[msg.ID]; ok {
					delete(c.outboundRequests, msg.ID)
					c.requestMu.Unlock()

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
				c.requestMu.Unlock()

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

				c.subscriptionsMu.RLock()
				if subscription, ok := c.subscriptions[sp.Subscription]; ok {
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
				c.subscriptionsMu.RUnlock()
			}
		}

	})

	// Writer
	g.Go(func() error {
		for {
			select {
			case r := <-c.chToBackend:
				// log.Printf("[SPAM] Writing %v", r)
				b, err := json.Marshal(&r)
				if err != nil {
					return errors.Wrap(err, "error marshalling request for backend")
				}

				// log.Printf("[SPAM] Writing %s", string(b))
				err = c.conn.WriteMessage(websocket.TextMessage, b)
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
			case s := <-c.chSubscriptionRequests:
				// log.Printf("[SPAM] Sub request %v", s)
				id := c.nextID()
				proxy := *s.request
				proxy.ID = id

				c.requestMu.Lock()
				c.subscriptonRequests[id] = s
				c.requestMu.Unlock()

				select {
				case <-ctx.Done():
					return ctx.Err()
				case c.chToBackend <- proxy:
					continue
				}

			// outbound requests
			case o := <-c.chOutboundRequests:
				// log.Printf("[SPAM] outbound request %v", *o)
				id := c.nextID()
				proxy := *o.request
				proxy.ID = id
				// log.Printf("[SPAM] outbound proxied request %v", proxy)

				c.requestMu.Lock()
				c.outboundRequests[id] = o
				c.requestMu.Unlock()

				select {
				case <-ctx.Done():
					return ctx.Err()
				case c.chToBackend <- proxy:
					continue
				}

			case <-ctx.Done():
				return nil
			}
		}
	})

	// Aborter
	g.Go(func() error {
		<-ctx.Done()
		log.Printf("[DEBUG] Context done, setting deadlines to now")
		_ = c.conn.SetReadDeadline(time.Now())
		_ = c.conn.SetWriteDeadline(time.Now())
		return nil
	})

	err := g.Wait()
	if err == nil {
		err = context.Canceled
	}

	// let's clean up all the remaining subscriptions
	c.subscriptionsMu.Lock()
	for id, sub := range c.subscriptions {
		sub.mu.Lock()
		sub.err = err
		sub.mu.Unlock()
		sub.cancel()
		delete(c.subscriptions, id)
	}
	c.subscriptionsMu.Unlock()

	_ = c.conn.Close()
}

func (c *connection) nextID() jsonrpc.ID {
	return jsonrpc.ID{
		Num: atomic.AddUint64(&c.counter, 1),
	}
}

func (c *connection) URL() string {
	return c.url
}

func (c *connection) Request(ctx context.Context, r *jsonrpc.Request) (*jsonrpc.RawResponse, error) {

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
	case c.chOutboundRequests <- &outbound:
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

func (c *connection) Subscribe(ctx context.Context, r *jsonrpc.Request) (Subscription, error) {
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
	case c.chSubscriptionRequests <- &start:
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

func (c *connection) BlockNumber(ctx context.Context) (uint64, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_blockNumber",
	}

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

func (c *connection) BlockByNumber(ctx context.Context, number uint64, full bool) (*eth.Block, error) {
	n := eth.QuantityFromUInt64(number)

	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getBlockByNumber",
		Params: jsonrpc.MustParams(&n, full),
	}

	// log.Printf("[SPAM] params: [%s, %s]", string(request.Params[0]), string(request.Params[1]))

	response, err := c.Request(ctx, &request)
	if err != nil {
		return nil, errors.Wrap(err, "could not make request")
	}

	return c.parseBlockResponse(response)
}

func (c *connection) parseBlockResponse(response *jsonrpc.RawResponse) (*eth.Block, error) {
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

func (c *connection) BlockByHash(ctx context.Context, hash string, full bool) (*eth.Block, error) {
	h, err := eth.NewHash(hash)
	if err != nil {
		return nil, errors.Wrap(err, "invalid hash")
	}

	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getBlockByHash",
		Params: jsonrpc.MustParams(h, full),
	}

	response, err := c.Request(ctx, &request)
	if err != nil {
		return nil, errors.Wrap(err, "could not make request")
	}

	return c.parseBlockResponse(response)
}

func (c *connection) TransactionByHash(ctx context.Context, hash string) (*eth.Transaction, error) {
	h, err := eth.NewHash(hash)
	if err != nil {
		return nil, errors.Wrap(err, "invalid hash")
	}

	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getTransactionByHash",
		Params: jsonrpc.MustParams(h),
	}

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

func (c *connection) NewHeads(ctx context.Context) (Subscription, error) {
	r := jsonrpc.Request{
		JSONRPC: "2.0",
		ID:      jsonrpc.ID{Str: "test", IsString: true},
		Method:  "eth_subscribe",
		Params:  jsonrpc.MustParams("newHeads"),
	}

	return c.Subscribe(ctx, &r)
}

// if full is set to true, includeTransactions will be set to true for subscription parameters
func (c *connection) NewPendingTransaction(ctx context.Context, full ...bool) (Subscription, error) {

	var r jsonrpc.Request
	if full != nil {
		r = jsonrpc.Request{
			JSONRPC: "2.0",
			ID:      jsonrpc.ID{Str: "pending", IsString: true},
			Method:  "eth_subscribe",
			Params:  jsonrpc.MustParams("newPendingTransactions", map[string]interface{}{"includeTransactions": full[0]}),
		}
	} else {
		r = jsonrpc.Request{
			JSONRPC: "2.0",
			ID:      jsonrpc.ID{Str: "pending", IsString: true},
			Method:  "eth_subscribe",
			Params:  jsonrpc.MustParams("newPendingTransactions"),
		}
	}
	return c.Subscribe(ctx, &r)
}

func (c *connection) TransactionReceipt(ctx context.Context, hash string) (*eth.TransactionReceipt, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getTransactionReceipt",
		Params: jsonrpc.MustParams(hash),
	}

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

func (c *connection) GetLogs(ctx context.Context, filter eth.LogFilter) ([]eth.Log, error) {
	request := jsonrpc.Request{
		ID:     jsonrpc.ID{Num: 1},
		Method: "eth_getLogs",
		Params: jsonrpc.MustParams(filter),
	}

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
