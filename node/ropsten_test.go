package node_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/INFURA/go-ethlibs/node"
)

func getRopstenClient(t *testing.T, ctx context.Context) node.Client {
	// These test require a ropsten websocket URL to test with, for example ws://localhost:8546 or wss://ropsten.infura.io/ws/v3/:YOUR_PROJECT_ID
	url := os.Getenv("ETHLIBS_TEST_ROPSTEN_WS_URL")
	if url == "" {
		t.Skip("ETHLIBS_TEST_ROPSTEN_WS_URL not set, skipping test.  Set to a valid websocket URL to execute this test.")
	}

	conn, err := node.NewClient(ctx, url)
	require.NoError(t, err, "creating websocket connection should not fail")
	return conn
}

func TestConnection_PendingNonceAt(t *testing.T) {
	ctx := context.Background()
	conn := getRopstenClient(t, ctx)

	pendingNonce, err := conn.PendingNonceAt(ctx, "0xed28874e52A12f0D42118653B0FBCee0ACFadC00", "latest")
	require.NoError(t, err)
	require.NotEmpty(t, pendingNonce, "pending nonce must not be nil")
}

func TestConnection_NetworkID(t *testing.T) {
	ctx := context.Background()
	conn := getRopstenClient(t, ctx)

	chainID, err := conn.NetworkID(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, chainID, "chain id must not be nil")
}

func TestConnection_FutureBlockByNumber(t *testing.T) {
	ctx := context.Background()
	conn := getRopstenClient(t, ctx)

	blockNumber, err := conn.BlockNumber(ctx)
	require.NoError(t, err)

	next, err := conn.BlockByNumber(ctx, blockNumber+1000, false)
	require.Nil(t, next, "future block should be nil")
	require.Error(t, err, "requesting a future block should return an error")
	require.Equal(t, node.ErrBlockNotFound, err)

	// get a the genesis block by number which should _not_ fail
	genesis, err := conn.BlockByNumber(ctx, 0, false)
	require.NoError(t, err, "requesting genesis block by number should not fail")
	require.NotNil(t, genesis, "genesis block must not be nil")
}

func TestConnection_InvalidBlockByHash(t *testing.T) {
	ctx := context.Background()
	conn := getRopstenClient(t, ctx)

	b, err := conn.BlockByHash(ctx, "invalid", false)
	require.Error(t, err, "requesting an invalid hash should return an error")
	require.Nil(t, b, "block from invalid hash should be nil")

	b, err = conn.BlockByHash(ctx, "0x1234", false)
	require.Error(t, err, "requesting an invalid hash should return an error")
	require.Nil(t, b, "block from invalid hash should be nil")

	b, err = conn.BlockByHash(ctx, "0x0badf00dbadf00dbadf00dbadf00dbadf00dbadf00dbadf00dbadf00dbadf00d", false)
	require.Error(t, err, "requesting a non-existent block should should return an error")
	require.Nil(t, b, "block from non-existent hash should be nil")
	require.Equal(t, node.ErrBlockNotFound, err)

	// get the genesis block which should _not_ fail
	b, err = conn.BlockByHash(ctx, "0x41941023680923e0fe4d74a34bdac8141f2540e3ae90623718e47d66d1ca4a2d", true)
	require.NoError(t, err, "genesis block hash should not return an error")
	require.NotNil(t, b, "genesis block should be retrievable by hash")
}

func TestConnection_InvalidTransactionByHash(t *testing.T) {
	ctx := context.Background()
	conn := getRopstenClient(t, ctx)

	tx, err := conn.TransactionByHash(ctx, "invalid")
	require.Error(t, err, "requesting an invalid hash should return an error")
	require.Nil(t, tx, "tx from invalid hash should be nil")

	tx, err = conn.TransactionByHash(ctx, "0x1234")
	require.Error(t, err, "requesting an invalid hash should return an error")
	require.Nil(t, tx, "tx from invalid hash should be nil")

	tx, err = conn.TransactionByHash(ctx, "0x0badf00dbadf00dbadf00dbadf00dbadf00dbadf00dbadf00dbadf00dbadf00d")
	require.Error(t, err, "requesting an non-existent hash should return an error")
	require.Nil(t, tx, "tx from non-existent hash should be nil")
	require.Equal(t, node.ErrTransactionNotFound, err)

	// get an early tx which should _not_ fail
	tx, err = conn.TransactionByHash(ctx, "0x230f6e1739286f9cbf768e34a9ff3d69a2a72b92c8c3383fbdf163035c695332")
	require.NoError(t, err, "early tx should not return an error")
	require.NotNil(t, tx, "early tx should be retrievable by hash")
}
