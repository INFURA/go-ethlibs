package noderpc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func (c *Client) BlockNumber(
	ctx context.Context,
	url string,
	headers map[string]string,
) (uint64, error) {
	const (
		method = "eth_blockNumber"
		empty  = uint64(0)
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

	var response BlockNumberResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("eth_blockNumber response: %v",body)

		return empty, fmt.Errorf("%w", err)
	}

	blockNumber, err := hexutil.DecodeUint64(response.Result)
	if err != nil {
		return empty, fmt.Errorf("%w", err)
	}

	return blockNumber, nil
}
