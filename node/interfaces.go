package node

import (
	"context"
	"math/big"

	"github.com/INFURA/go-ethlibs/eth"
	"github.com/INFURA/go-ethlibs/jsonrpc"
)

type Requester interface {
	// Request method can be used to send JSONRPC requests and receive JSONRPC responses
	Request(ctx context.Context, r *jsonrpc.Request) (*jsonrpc.RawResponse, error)
}

type Subscriber interface {
	// Subscribe method can be used to subscribe via eth_subscribe
	Subscribe(ctx context.Context, r *jsonrpc.Request) (Subscription, error)
}

// Client represents a connection to an ethereum node
type Client interface {
	Requester
	Subscriber

	// URL returns the backend URL we are connected to
	URL() string

	// BlockNumber returns the current block number at head
	BlockNumber(ctx context.Context) (uint64, error)

	// NetworkID returns the chain id
	NetworkID(ctx context.Context) (*big.Int, error)

	// EstimateGas returns the estimate gas
	EstimateGas(ctx context.Context, msg eth.Transaction) (uint64, error)

	// SuggestGasTipCap (EIP1559) returns the suggested tip for block
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)

	// SuggestGasPrice (Legacy) returns the suggested gas price
	SuggestGasPrice(ctx context.Context) (*big.Int, error)

	// PendingNonceAt get the pending nonce for public address
	PendingNonceAt(ctx context.Context, address string, block string) (uint64, error)

	// BlockByNumber can be used to get a block by its number
	BlockByNumber(ctx context.Context, number uint64, full bool) (*eth.Block, error)

	// BlockByNumberOrTag can be used to get a block by its number or tag (e.g. latest)
	BlockByNumberOrTag(ctx context.Context, numberOrTag eth.BlockNumberOrTag, full bool) (*eth.Block, error)

	// BlockByHash can be used to get a block by its hash
	BlockByHash(ctx context.Context, hash string, full bool) (*eth.Block, error)

	// TransactionByHash can be used to get transaction by its hash
	TransactionByHash(ctx context.Context, hash string) (*eth.Transaction, error)

	// SubscribeNewHeads initiates a subscription for newHead events
	SubscribeNewHeads(ctx context.Context) (Subscription, error)

	// SubscribeNewPendingTransactions initiates a subscription for newPendingTransaction events
	SubscribeNewPendingTransactions(ctx context.Context) (Subscription, error)

	// TransactionReceipt can be used to get a TransactionReceipt for a particular transaction
	TransactionReceipt(ctx context.Context, hash string) (*eth.TransactionReceipt, error)

	// Logs returns an array of Logs matching the passed in filter
	Logs(ctx context.Context, filter eth.LogFilter) ([]eth.Log, error)

	// IsBidirectional returns true if the under laying transport supports bidirectional features such as subscriptions
	IsBidirectional() bool
}

type Subscription interface {
	Response() *jsonrpc.RawResponse
	ID() string
	Ch() <-chan *jsonrpc.Notification
	Unsubscribe(ctx context.Context) error
}
