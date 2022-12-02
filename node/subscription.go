package node

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/ConsenSys/go-ethlibs/jsonrpc"
)

type subscription struct {
	response        *jsonrpc.RawResponse
	subscriptionID  string
	notificationsCh chan *jsonrpc.Notification
	dispatchCh      chan *jsonrpc.Notification
	signalCh        chan struct{}
	stoppedCh       chan struct{}
	conn            Requester
}

func (s *subscription) Response() *jsonrpc.RawResponse {
	return s.response
}

func (s *subscription) ID() string {
	return s.subscriptionID
}

func (s *subscription) Ch() <-chan *jsonrpc.Notification {
	return s.notificationsCh
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

	response, err := s.conn.Request(ctx, &request)
	if err != nil {
		return errors.Wrap(err, "unsubscribe failed")
	}

	if response.Error != nil {
		return errors.Errorf("%s", string(*response.Error))
	}

	return nil
}

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

	// start a goroutine to move work from dispatchCh to notificationsCh
	// we have to have this intermediate channel and the two
	// signalling channels to make sure that everything
	// is shut down in the proper order and no one tries writing
	// to a closed channel.  See for example variant #5 on https://go101.org/article/channel-closing.html
	go func() {
		defer func() {
			// Close the stopped channel so it signals to everyone that this goroutine
			// is no longer running, which will cause .dispatch and .stop to return without blocking
			close(s.stoppedCh)

			// then close the notifications channel so any consumers of
			// subscription.Ch() are unblocked
			close(s.notificationsCh)
		}()

		for {
			select {
			case <-s.signalCh:
				// we've been signalled to stop, we can end this goroutine
				return
			case n := <-s.dispatchCh:
				// we have a notification to dispatch to the external channel
				select {
				case s.notificationsCh <- n:
					// we successfully dispatched the notification
				case <-s.signalCh:
					// we were told to stop while trying to write the notification,
					// we can end this goroutine
					return
				}
			}
		}
	}()

	return &s
}

func (s *subscription) dispatch(ctx context.Context, n jsonrpc.Notification) {
	// we've been given a notification that needs to be dispatched to notificationsCh
	// however we can't do it directly here, since we only want one goroutine sending
	// notifications on this channel, so that it can be stopped cleanly.
	// Instead, we'll write it to an intermediate work/dispatch channel.
	//
	// Important note: the `n` argument to this function is intentionally a by-value copy
	// of the full struct rahter than a pointer.  This is to ensure that this function owns
	// the pointer that is written to the dispatch channel.

	select {
	case <-ctx.Done():
		// the parent context .dispatch was called ended, we can go ahead and give up
		// on dispatching this notification
		return
	case s.dispatchCh <- &n:
		// the notification is now in the internal dispatch channel and will be processed
		// by the goroutine created in newSubscription above
		return
	case <-s.stoppedCh:
		// this subscription has been stopped so we can abandon any writes to it
		return
	}
}

func (s *subscription) stop(ctx context.Context) {
	select {
	case <-ctx.Done():
		// the calling context has ended, presumably because the client is shutting down
		return
	case <-s.stoppedCh:
		// we've been stopped already, presumably by another caller of .stop()
		return
	case s.signalCh <- struct{}{}:
		// we successfully signalled to the newSubscription goroutine our desire to stop
		return
	}
}
