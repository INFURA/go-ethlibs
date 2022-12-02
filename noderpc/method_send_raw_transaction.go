package noderpc

import (
	"context"
	"encoding/json"
	"fmt"
)

func (c *Client) SendRawTransaction(ctx context.Context,
	url string,
	headers map[string]string,
	rawTransaction string,
) (SendRawTransactionError, error) {
	const (
		method string = "eth_sendRawTransaction"
	)

	empty := SendRawTransactionError{} //nolint:exhaustivestruct,exhaustruct

	body, err := c.Call(
		ctx,
		url,
		method,
		[]interface{}{
			rawTransaction,
		},
		headers,
	)
	if err != nil {
		return empty, fmt.Errorf("%w", err)
	}

	var response SendRawTransactionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return empty, fmt.Errorf("%w", err)
	}

	return response.Error, nil
}
