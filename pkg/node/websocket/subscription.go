package websocket

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"sync"

	"github.com/INFURA/eth/pkg/jsonrpc"
)

type Subscription interface {
	Response() *jsonrpc.RawResponse
	ID() string
	Ch() chan *jsonrpc.Notification
	Unsubscribe(ctx context.Context) error
	Done() <-chan struct{}
	Err() error
}

type subscription struct {
	response       *jsonrpc.RawResponse
	subscriptionID string
	ch             chan *jsonrpc.Notification

	backend *backend
	ctx     context.Context
	cancel  context.CancelFunc
	err     error
	mu      sync.RWMutex
}

func (s *subscription) Response() *jsonrpc.RawResponse {
	return s.response
}

func (s *subscription) ID() string {
	return s.subscriptionID
}

func (s *subscription) Ch() chan *jsonrpc.Notification {
	return s.ch
}

type SubscriptionParams struct {
	Subscription string          `json:"subscription"`
	Result       json.RawMessage `json:"result"`
}

func (s *subscription) Unsubscribe(ctx context.Context) error {
	request := jsonrpc.Request{
		ID: jsonrpc.ID{
			Str: s.subscriptionID,
		},
		Method: "eth_unsubscribe",
		Params: jsonrpc.MustParams(s.subscriptionID),
	}

	response, err := s.backend.Request(ctx, &request)
	if err != nil {
		return errors.Wrap(err, "unsubscribe failed")
	}

	if response.Error != nil {
		return errors.Errorf("%v", response.Error)
	}

	s.cancel()
	return nil
}

func (s *subscription) Done() <-chan struct{} {
	return s.ctx.Done()
}

func (s *subscription) Err() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.err != nil {
		return s.err
	}
	return s.ctx.Err()
}
