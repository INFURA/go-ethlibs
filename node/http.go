package node

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/INFURA/go-ethlibs/jsonrpc"
	"github.com/pkg/errors"
)

func newHTTPTransport(ctx context.Context, parsedURL *url.URL, header http.Header) (transport, error) {
	return &httpTransport{
		header: header,
		rawURL: parsedURL.String(),
	}, nil
}

type httpTransport struct {
	header http.Header
	rawURL string
	client *http.Client
	once   sync.Once
}

func (t *httpTransport) Request(ctx context.Context, r *jsonrpc.Request) (*jsonrpc.RawResponse, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode request json")
	}

	body, err := t.dispatchBytes(ctx, b)
	if err != nil {
		return nil, errors.Wrap(err, "could not dispatch request")
	}

	jr := jsonrpc.RawResponse{}
	err = json.Unmarshal(body, &jr)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode response json")
	}

	return &jr, nil
}

func (t *httpTransport) Subscribe(ctx context.Context, r *jsonrpc.Request) (Subscription, error) {
	return nil, errors.New("subscriptions not supported over HTTP")
}

func (t *httpTransport) IsBidirectional() bool {
	return false
}

func (t *httpTransport) dispatchBytes(ctx context.Context, input []byte) ([]byte, error) {
	t.once.Do(func() {
		// Since this client is only ever used to access a single endpoint,
		// we allow all the idle connections to point that host
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.MaxIdleConnsPerHost = tr.MaxIdleConns
		t.client = &http.Client{
			Timeout:   120 * time.Second,
			Transport: tr,
		}
	})

	r, err := http.NewRequest(http.MethodPost, t.rawURL, bytes.NewReader(input))
	if err != nil {
		return nil, errors.Wrap(err, "could not create http.Request")
	}

	r = r.WithContext(ctx)
	r.Header = t.header
	r.Header.Add("Content-Type", "application/json")

	resp, err := t.client.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "error in client.Do")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "error reading body")
	}

	return body, nil
}
