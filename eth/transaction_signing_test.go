package eth_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/INFURA/go-ethlibs/eth"
)

func TestSignTransaction(t *testing.T) {
	chainId := eth.QuantityFromInt64(1)
	tx := eth.Transaction{
		Nonce:    eth.QuantityFromUInt64(0),
		GasPrice: eth.QuantityFromUInt64(21488430592),
		Gas:      eth.QuantityFromUInt64(90000),
		To:       eth.MustAddress("0xc149Be1bcDFa69a94384b46A1F91350E5f81c1AB"),
		Value:    eth.QuantityFromUInt64(950000000000000000),
		Input:    *eth.MustData("0x"),
	}

	// This purposefully uses the already highly compromised keypair from the go-ethereum book:
	// https://goethereumbook.org/transfer-eth/
	// privateKey = fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19
	signed, err := tx.Sign("0xfad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19", chainId)
	require.NoError(t, err)

	// check tx can be restored from rawtx
	tx2 := eth.Transaction{}
	err = tx2.FromRaw(signed.String())
	require.NoError(t, err)
	require.Equal(t, tx2.Nonce.UInt64(), uint64(0))
	require.Equal(t, tx2.GasPrice, eth.QuantityFromInt64(21488430592))
	require.Equal(t, tx2.Gas, eth.QuantityFromInt64(90000))
	require.Equal(t, tx2.To.String(), "0xc149Be1bcDFa69a94384b46A1F91350E5f81c1AB")
	require.Equal(t, tx2.Value, eth.QuantityFromInt64(950000000000000000))
}

func TestSignTransaction2(t *testing.T) {
	chainId := eth.QuantityFromInt64(1)
	tx := eth.Transaction{
		Nonce:    eth.QuantityFromUInt64(146),
		GasPrice: eth.QuantityFromUInt64(3000000000),
		Gas:      eth.QuantityFromUInt64(22000),
		To:       eth.MustAddress("0x43700db832E9Ac990D36d6279A846608643c904E"),
		Value:    eth.QuantityFromUInt64(1000000000),
		Input:    *eth.MustData("0x"),
	}

	signed, err := tx.Sign("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19", chainId)
	require.NoError(t, err)

	// check tx can be restored from rawtx
	tx2 := eth.Transaction{}
	err = tx2.FromRaw(signed.String())
	require.NoError(t, err)
	require.Equal(t, tx2.Nonce, eth.QuantityFromInt64(146))
	require.Equal(t, tx2.GasPrice, eth.QuantityFromInt64(3000000000))
	require.Equal(t, tx2.Gas, eth.QuantityFromInt64(22000))
	require.Equal(t, tx2.To.String(), "0x43700db832E9Ac990D36d6279A846608643c904E")
	require.Equal(t, tx2.Value, eth.QuantityFromInt64(1000000000))
}

// compares signed output created in python script
// signed = w3.eth.account.signTransaction(transaction, pKey)
// where pKey = `fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19`
func TestSignTransaction3(t *testing.T) {
	chainId := eth.QuantityFromInt64(1)
	raw := eth.MustData("0xf868819284b2d05e008255f09443700db832e9ac990d36d6279a846608643c904e843b9aca008026a0444f6cd588830bc975643241e6df545dccf5815c00ee8bde4e686722761b8954a06abec148bf44975c6ed6336cba57a9f5101d1cb5c199a12567d65de2ea8d7d43")
	tx := eth.Transaction{
		Nonce:    eth.QuantityFromUInt64(146),
		GasPrice: eth.QuantityFromUInt64(3000000000),
		Gas:      eth.QuantityFromUInt64(22000),
		To:       eth.MustAddress("0x43700db832E9Ac990D36d6279A846608643c904E"),
		Value:    eth.QuantityFromUInt64(1000000000),
		Input:    *eth.MustData("0x"),
	}

	signed, err := tx.Sign("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19", chainId)
	require.NoError(t, err)

	require.Equal(t, raw.String(), signed.String())
}
