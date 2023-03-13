package node_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/INFURA/go-ethlibs/eth"
	"github.com/INFURA/go-ethlibs/node"
)

func getGoerliClient(t *testing.T, ctx context.Context) node.Client {
	// These test require a ropsten websocket URL to test with, for example ws://localhost:8546 or wss://ropsten.infura.io/ws/v3/:YOUR_PROJECT_ID
	url := os.Getenv("ETHLIBS_TEST_GOERLI_WS_URL")
	if url == "" {
		t.Skip("ETHLIBS_TEST_GOERLI_WS_URL not set, skipping test.  Set to a valid websocket URL to execute this test.")
	}

	conn, err := node.NewClient(ctx, url)
	require.NoError(t, err, "creating websocket connection should not fail")
	return conn
}

func TestConnection_GetTransactionCount(t *testing.T) {
	ctx := context.Background()
	conn := getGoerliClient(t, ctx)

	// Checks the current pending nonce for account can be retrieved
	blockNum1 := eth.MustBlockNumberOrTag("latest")
	pendingNonce1, err := conn.GetTransactionCount(ctx, "0xed28874e52A12f0D42118653B0FBCee0ACFadC00", *blockNum1)
	require.NoError(t, err)
	require.NotEmpty(t, pendingNonce1, "pending nonce must not be nil")

	// Should catch failure since it is looking for a nonce of a future block
	blockNum2 := eth.MustBlockNumberOrTag("0x7654321")
	pendingNonce2, err := conn.GetTransactionCount(ctx, "0xed28874e52A12f0D42118653B0FBCee0ACFadC00", *blockNum2)
	require.Error(t, err)
	require.Empty(t, pendingNonce2, "pending nonce must not exist since it is a future block")
}

func TestConnection_EstimateGas(t *testing.T) {
	ctx := context.Background()
	conn := getGoerliClient(t, ctx)

	from := eth.MustAddress("0xed28874e52A12f0D42118653B0FBCee0ACFadC00")
	tx := eth.Transaction{
		Nonce:    eth.QuantityFromUInt64(146),
		GasPrice: eth.OptionalQuantityFromInt(3000000000),
		Gas:      eth.QuantityFromUInt64(22000),
		To:       eth.MustAddress("0x43700db832E9Ac990D36d6279A846608643c904E"),
		Value:    *eth.OptionalQuantityFromInt(100),
		From:     *from,
	}

	gas, err := conn.EstimateGas(ctx, tx)
	require.NoError(t, err)
	require.NotEqual(t, gas, 0, "estimate gas cannot be equal to zero.")
}

func TestConnection_MaxPriorityFeePerGas(t *testing.T) {
	ctx := context.Background()
	conn := getGoerliClient(t, ctx)

	fee, err := conn.MaxPriorityFeePerGas(ctx)
	require.NoError(t, err)
	require.NotEqual(t, fee, 0, "fee cannot be equal to 0")
}

func TestConnection_GasPrice(t *testing.T) {
	ctx := context.Background()
	conn := getGoerliClient(t, ctx)

	gasPrice, err := conn.GasPrice(ctx)
	require.NoError(t, err)
	require.NotEqual(t, gasPrice, 0, "gas price cannot be equal to 0")
}

func TestConnection_NetVersion(t *testing.T) {
	ctx := context.Background()
	conn := getGoerliClient(t, ctx)

	netVersion, err := conn.NetVersion(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, netVersion, "net version id must not be nil")
}

func TestConnection_ChainId(t *testing.T) {
	ctx := context.Background()
	conn := getGoerliClient(t, ctx)

	chainId, err := conn.ChainId(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, chainId, "chain id must not be nil")
}

func TestConnection_SendRawTransactionInValidEmpty(t *testing.T) {
	ctx := context.Background()
	conn := getGoerliClient(t, ctx)

	txHash, err := conn.SendRawTransaction(ctx, "0x0")
	require.Error(t, err)
	require.Empty(t, txHash, "txHash must be nil")
}

func TestConnection_SendRawTransactionInValidOldNonce(t *testing.T) {
	ctx := context.Background()
	conn := getGoerliClient(t, ctx)

	data := eth.MustData("0xf86e0185174876e8008252089460c063d3f3b744e2d153fcbe66a068b09109cf1b865af3107a400084baadf00d2ea0b4d9e2edbd2a2d9a38cf0415f9d03849e6a6f2de8562d7cd74eda89397882030a056edb455e9ffa07ad22f8b06f9065564911f796a026e1b2177ecaad995198aaa")
	txHash, err := conn.SendRawTransaction(ctx, data.String())
	require.Error(t, err)
	require.Equal(t, err.Error(), "{\"code\":-32000,\"message\":\"nonce too low\"}")
	require.Empty(t, txHash, "txHash must be nil")
}

func TestConnection_FutureBlockByNumber(t *testing.T) {
	ctx := context.Background()
	conn := getGoerliClient(t, ctx)

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
	conn := getGoerliClient(t, ctx)

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
	// https://goerli.etherscan.io/block/8000000
	b, err = conn.BlockByHash(ctx, "0x2ae83825ac6b2a2b2509da8617cf31072a5628e9a818f177316f4f4bcdfafd06", true)
	require.NoError(t, err, "genesis block hash should not return an error")
	require.NotNil(t, b, "genesis block should be retrievable by hash")
}

func TestConnection_InvalidTransactionByHash(t *testing.T) {
	ctx := context.Background()
	conn := getGoerliClient(t, ctx)

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
	// https://goerli.etherscan.io/tx/0x752ca2e3175c0dfd8b8612abcd2dac3134445f29e764d33645726cbcd57aefd1
	tx, err = conn.TransactionByHash(ctx, "0x752ca2e3175c0dfd8b8612abcd2dac3134445f29e764d33645726cbcd57aefd1")
	require.NoError(t, err, "early tx should not return an error")
	require.NotNil(t, tx, "early tx should be retrievable by hash")
}
