package node

import (
	"bufio"
	"context"
	"net"
	"net/url"

	"github.com/pkg/errors"
)

func newIPCTransport(ctx context.Context, parsedURL *url.URL) (*ipcTransport, error) {
	conn, err := net.Dial("unix", parsedURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "could not connect over IPC")
	}
	scanner := bufio.NewScanner(conn)
	readMessage := func() (payload []byte, err error) {
		if !scanner.Scan() {
			return nil, ctx.Err()
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}

		payload = []byte(scanner.Text())
		err = nil
		return
	}

	writeMessage := func(payload []byte) error {
		_, err := conn.Write(payload)
		return err
	}

	t := ipcTransport{
		loopingTransport: newLoopingTransport(ctx, conn, readMessage, writeMessage),
	}

	return &t, nil
}

type ipcTransport struct {
	loopingTransport
}
