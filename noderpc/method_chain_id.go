package noderpc

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) ChainID(
	ctx context.Context,
	url string,
	headers map[string]string,
) (string, error) {
	const (
		method string = "eth_chainId"
		empty  string = ""
	)

	body, err := c.Call(
		ctx,
		url,
		method,
		[]interface{}{},
		headers,
	)
	if err != nil {
		return empty, fmt.Errorf("%w", err)
	}

	var response ChainIDResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return empty, fmt.Errorf("%w", err)
	}

	return response.Result, nil
}
