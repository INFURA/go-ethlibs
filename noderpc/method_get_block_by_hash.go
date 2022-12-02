package noderpc

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) GetBlockByHash(
	ctx context.Context,
	url string,
	headers map[string]string,
	blockHash string,
) (BlockResult, error) {
	const (
		method string = "eth_getBlockByHash"
	)

	empty := BlockResult{} //nolint:exhaustivestruct,exhaustruct

	body, err := c.Call(
		ctx,
		url,
		method,
		[]interface{}{
			blockHash, false,
		},
		headers,
	)
	if err != nil {
		return empty, fmt.Errorf("%w", err)
	}

	var response BlockByHashResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return empty, fmt.Errorf("%w", err)
	}

	return response.Result, nil
}
