package noderpc

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) Syncing(
	ctx context.Context,
	url string,
	headers map[string]string,
) (bool, error) {
	const method string = "eth_syncing"

	body, err := c.Call(
		ctx,
		url,
		method,
		[]interface{}{},
		headers,
	)
	if err != nil {
		return true, fmt.Errorf("%w", err)
	}

	var response SyncingResponse
	if err = json.Unmarshal(body, &response); err != nil {
		var responseSyncing SyncingInProgressResponse
		if err = json.Unmarshal(body, &responseSyncing); err != nil {
			return true, fmt.Errorf("%w %s", err, body)
		}

		//nolint:goerr113
		return true, fmt.Errorf("syncing - currentBlock: %s, highestBlock: %s, "+
			"knownStates: %s, pulledStates: %s, startingBlock: %s",
			responseSyncing.Result.CurrentBlock, responseSyncing.Result.HighestBlock,
			responseSyncing.Result.KnownStates, responseSyncing.Result.PulledStates, responseSyncing.Result.StartingBlock)
	}

	return response.Result, nil
}
