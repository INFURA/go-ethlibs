package eth_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/INFURA/go-ethlibs/eth"
)

func TestSignTransaction(t *testing.T) {
	data := "0x"
	//tx, err := eth.NewTransaction(5, 21488430592, 90000, "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed", 1, data)
	tx, err := eth.NewTransaction(0, 21488430592, 90000, "0xc149Be1bcDFa69a94384b46A1F91350E5f81c1AB", 950000000000000000, data)
	require.NoError(t, err)
	require.Equal(t, tx.Nonce.UInt64(), uint64(0))
	require.Equal(t, tx.GasPrice, eth.QuantityFromInt64(21488430592))
	require.Equal(t, tx.Gas, eth.QuantityFromInt64(90000))
	require.Equal(t, tx.To.String(), "0xc149Be1bcDFa69a94384b46A1F91350E5f81c1AB")
	require.Equal(t, tx.Value, eth.QuantityFromInt64(950000000000000000))

	// This purposefully uses the already highly compromised keypair from the go-ethereum book:
	// https://goethereumbook.org/transfer-eth/
	// privateKey = fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19

	signed, err := tx.Sign("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19", 1)
	require.NoError(t, err)

	// check tx can be restored from rawtx
	tx2 := eth.Transaction{}
	err = tx2.FromRaw(signed)
	require.NoError(t, err)
	require.Equal(t, tx2.Nonce.UInt64(), uint64(0))
	require.Equal(t, tx2.GasPrice, eth.QuantityFromInt64(21488430592))
	require.Equal(t, tx2.Gas, eth.QuantityFromInt64(90000))
	require.Equal(t, tx2.To.String(), "0xc149Be1bcDFa69a94384b46A1F91350E5f81c1AB")
	require.Equal(t, tx2.Value, eth.QuantityFromInt64(950000000000000000))
}

func TestSignTransaction2(t *testing.T) {
	data := "0x"
	tx, err := eth.NewTransaction(146, 3000000000, 22000, "0x43700db832E9Ac990D36d6279A846608643c904E", 1000000000, data)
	require.NoError(t, err)
	require.Equal(t, tx.Nonce, eth.QuantityFromInt64(146))
	require.Equal(t, tx.GasPrice, eth.QuantityFromInt64(3000000000))
	require.Equal(t, tx.Gas, eth.QuantityFromInt64(22000))
	require.Equal(t, tx.To.String(), "0x43700db832E9Ac990D36d6279A846608643c904E")
	require.Equal(t, tx.Value, eth.QuantityFromInt64(1000000000))

	signed, err := tx.Sign("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19", 1)
	require.NoError(t, err)

	// check tx can be restored from rawtx
	tx2 := eth.Transaction{}
	err = tx2.FromRaw(signed)
	require.NoError(t, err)
	require.Equal(t, tx2.Nonce, eth.QuantityFromInt64(146))
	require.Equal(t, tx2.GasPrice, eth.QuantityFromInt64(3000000000))
	require.Equal(t, tx2.Gas, eth.QuantityFromInt64(22000))
	require.Equal(t, tx2.To.String(), "0x43700db832E9Ac990D36d6279A846608643c904E")
	require.Equal(t, tx2.Value, eth.QuantityFromInt64(1000000000))

}

// compares signed output created in python script
// signed = w3.eth.account.signTransaction(transaction, MY_METAMASK_KEY)
// Need to figure out to get this test working
// The signature.V value for python signer is 28 while we are using 27 for our signer so the hash values are off by 1 for V
// Been manually changing the 'V' value to 28 and running to the test to make sure we are getting correct rawtx
func TestSignTransaction3(t *testing.T) {
    pythonRawTx := "0xf868819284b2d05e008255f09443700db832e9ac990d36d6279a846608643c904e843b9aca008026a0444f6cd588830bc975643241e6df545dccf5815c00ee8bde4e686722761b8954a06abec148bf44975c6ed6336cba57a9f5101d1cb5c199a12567d65de2ea8d7d43"
    data := "0x"
    tx, err := eth.NewTransaction(146, 3000000000, 22000, "0x43700db832E9Ac990D36d6279A846608643c904E", 1000000000, data)
    require.NoError(t, err)
    require.Equal(t, tx.Nonce, eth.QuantityFromInt64(146))
    require.Equal(t, tx.GasPrice, eth.QuantityFromInt64(3000000000))
    require.Equal(t, tx.Gas, eth.QuantityFromInt64(22000))
    require.Equal(t, tx.To.String(), "0x43700db832E9Ac990D36d6279A846608643c904E")
    require.Equal(t, tx.Value, eth.QuantityFromInt64(1000000000))

    signed, err := tx.Sign("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19", 1)
    require.NoError(t, err)

    pythonRawTx = strings.ToLower(pythonRawTx)
    signed = strings.ToLower(signed)
    require.Equal(t, signed, pythonRawTx)
}

