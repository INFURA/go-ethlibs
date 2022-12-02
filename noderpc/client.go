package noderpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// New create an instance of Client.
func New(httpClient *http.Client) *Client {
	return &Client{
		http:                httpClient,
		InfuraProjectID:     "",
		InfuraProjectSecret: "",
	}
}

// Call invokes a JSON-RPC method.
func (c *Client) Call(
	ctx context.Context,
	url string,
	method string,
	params interface{},
	headers map[string]string,
) ([]byte, error) {
	reqBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      1,
	})
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		url,
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if c.InfuraProjectID != "" && c.InfuraProjectSecret != "" {
		req.SetBasicAuth(c.InfuraProjectID, c.InfuraProjectSecret)
	}

	req.Header.Set("Content-Type", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return respBody, nil
}
