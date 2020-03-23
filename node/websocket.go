package node

import (
	"context"
	"io/ioutil"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type websocketTransport struct {
	*loopingTransport
}

// newWebsocketTransport creates a Connection to the passed in URL.  Use the supplied Context to shutdown the connection by
// cancelling or otherwise aborting the context.
func newWebsocketTransport(ctx context.Context, addr *url.URL) (transport, error) {
	wsConn, _, err := websocket.DefaultDialer.DialContext(ctx, addr.String(), nil)
	if err != nil {
		return nil, err
	}

	readMessage := func() (payload []byte, err error) {
		typ, r, err := wsConn.NextReader()
		if err != nil {
			return nil, errors.Wrap(err, "error reading from backend websocket connection")
		}

		if typ != websocket.TextMessage {
			return nil, nil
		}

		payload, err = ioutil.ReadAll(r)
		if err != nil {
			return nil, errors.Wrap(err, "error reading from backend websocket connection")
		}

		return payload, err
	}

	writeMessage := func(payload []byte) error {
		err := wsConn.WriteMessage(websocket.TextMessage, payload)
		return err
	}

	t := websocketTransport{
		loopingTransport: newLoopingTransport(ctx, wsConn, readMessage, writeMessage),
	}

	return &t, nil
}
