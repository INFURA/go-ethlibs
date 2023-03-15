package node

import (
	"context"
	"io/ioutil"
	"net/url"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type websocketTransport struct {
	*loopingTransport
}

// newWebsocketTransport creates a Connection to the passed in URL.  Use the supplied Context to shutdown the connection by
// cancelling or otherwise aborting the context.
func newWebsocketTransport(ctx context.Context, addr *url.URL, requestHeader http.Header) (transport, error) {
    // if getEnv INFURA_API_KEY && INFURA_PROJECT_ID then create Authorization header base64(INFURA_PROJECT_ID:INFURA_API_KEY)

	wsConn, _, err := websocket.DefaultDialer.DialContext(ctx, addr.String(), requestHeader)
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
