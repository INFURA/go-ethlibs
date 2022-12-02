package noderpc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func (c *Client) GetBlockByNumber(
	ctx context.Context,
	url string,
	headers map[string]string,
	blockNumber uint64,
) (BlockResult, error) {
	const (
		method string = "eth_getBlockByNumber"
	)

	empty := BlockResult{} //nolint:exhaustivestruct,exhaustruct

	body, err := c.Call(
		ctx,
		url,
		method,
		[]interface{}{hexutil.EncodeUint64(blockNumber), false},
		headers,
	)
	if err != nil {
		return empty, fmt.Errorf("%w", err)
	}

	var response BlockByNumberResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return empty, fmt.Errorf("%w", err)
	}

	response.Result.BlockNumberUI64, err = hexutil.DecodeUint64(response.Result.BlockNumber)
	if err != nil {
		return empty, fmt.Errorf("%w", err)
	}

	return response.Result, nil
}
