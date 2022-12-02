package node

import (
	"context"

	"github.com/pkg/errors"

	"github.com/ConsenSys/go-ethlibs/jsonrpc"
)

type customTransport struct {
	requester  Requester
	subscriber Subscriber
}

func (t *customTransport) Request(ctx context.Context, r *jsonrpc.Request) (*jsonrpc.RawResponse, error) {
	return t.requester.Request(ctx, r)
}

func (t *customTransport) Subscribe(ctx context.Context, r *jsonrpc.Request) (Subscription, error) {
	if t.subscriber == nil {
		return nil, errors.New("subscriptions not supported over this transport")
	}

	return t.subscriber.Subscribe(ctx, r)
}

func (t *customTransport) IsBidirectional() bool {
	return t.subscriber != nil
}

func newCustomTransport(requester Requester, subscriber Subscriber) (*customTransport, error) {
	t := customTransport{
		requester:  requester,
		subscriber: subscriber,
	}

	return &t, nil
}
