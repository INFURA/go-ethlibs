package noderpc

import (
	"net/http"
)

// Client represents a Client.
type Client struct {
	http                *http.Client
	InfuraProjectID     string
	InfuraProjectSecret string
}

// BlockResult represents the result struct.
type BlockResult struct {
	BlockNumberUI64 uint64
	BlockNumber     string `json:"number" validate:"required"`
	BlockHash       string `json:"hash" validate:"required"`
}

// BlockByHashResponse represents the response struct.
type BlockByHashResponse struct {
	Result BlockResult `json:"result" validate:"required"`
}

// BlockByNumberResponse represents the response struct.
type BlockByNumberResponse struct {
	Result BlockResult `json:"result" validate:"required"`
}

// BlockNumberResponse represents the response struct.
type BlockNumberResponse struct {
	Result string `json:"result" validate:"required"`
}

// ChainIDResponse represents the response struct.
type ChainIDResponse struct {
	Result string `json:"result" validate:"required"`
}

// SyncingResponse represents the response struct.
type SyncingResponse struct {
	Result bool `json:"result" validate:"required"`
}

// SyncingInProgressResponse represents the response when is not sync.
type SyncingInProgressResponse struct {
	Result SyncingInProgressDetailsResponse `json:"result" validate:"required"`
}
type SyncingInProgressDetailsResponse struct {
	CurrentBlock  string `json:"currentBlock" `
	HighestBlock  string `json:"highestBlock" `
	KnownStates   string `json:"knownStates" `
	PulledStates  string `json:"pulledStates" `
	StartingBlock string `json:"startingBlock" `
}

type SendRawTransactionError struct {
	Code    int    `json:"code" `
	Message string `json:"message" `
}

type SendRawTransactionResponse struct {
	Error SendRawTransactionError `json:"error" `
}
