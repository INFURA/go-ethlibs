package eth_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/INFURA/go-ethlibs/eth"
)

func TestTransaction_Sign(t *testing.T) {
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
	require.Equal(t, tx.Hash.String(), tx2.Hash.String())
}

func TestTransaction_Sign_2(t *testing.T) {
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
	require.Equal(t, tx.Hash.String(), tx2.Hash.String())
}

// compares signed output created in python script
// signed = w3.eth.account.signTransaction(transaction, pKey)
// where pKey = `fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19`
func TestTransaction_Sign_3(t *testing.T) {
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

func TestTransaction_Sign_EIP2930(t *testing.T) {
	chainId := eth.QuantityFromInt64(0x796f6c6f763378)

	// The following transaction was built using the geth console on yolov3x:
	/*
			> eth.fillTransaction({value: 1, from: "0x96216849c49358b10257cb55b28ea603c874b05e", to: "0xdf0a88b2b68c673713a8ec826003676f272e3573", accessList: [{"address": "0x0000000000000000000000000000000000001337","storageKeys": ["0x0000000000000000000000000000000000000000000000000000000000000000"]}] })
			{
			  raw: "0x01f86587796f6c6f76337880843b9aca008262d494df0a88b2b68c673713a8ec826003676f272e35730180f838f7940000000000000000000000000000000000001337e1a00000000000000000000000000000000000000000000000000000000000000000808080",
			  tx: {
			    accessList: [{
			        address: "0x0000000000000000000000000000000000001337",
			        storageKeys: [...]
			    }],
			    chainId: "0x796f6c6f763378",
			    gas: "0x62d4",
			    gasPrice: "0x3b9aca00",
			    hash: "0xf80c0c4c7e02360cb5dc4ef06fe619777d4e328504f107ae6b03e469f1a7b4de",
			    input: "0x",
			    nonce: "0x0",
			    r: "0x0",
			    s: "0x0",
			    to: "0xdf0a88b2b68c673713a8ec826003676f272e3573",
			    type: "0x1",
			    v: "0x0",
			    value: "0x1"
			  }
			}
		// And then signed via:
			> eth.signTransaction({value: 1, from: "0x96216849c49358b10257cb55b28ea603c874b05e", to: "0xdf0a88b2b68c673713a8ec826003676f272e3573", accessList: [{"address": "0x0000000000000000000000000000000000001337","storageKeys": ["0x0000000000000000000000000000000000000000000000000000000000000000"]}], gas: 0x62d4, gasPrice: 0x3b9aca00, nonce: 0x0 })
			{
			  raw: "0x01f8a587796f6c6f76337880843b9aca008262d494df0a88b2b68c673713a8ec826003676f272e35730180f838f7940000000000000000000000000000000000001337e1a0000000000000000000000000000000000000000000000000000000000000000080a0294ac94077b35057971e6b4b06dfdf55a6fbed819133a6c1d31e187f1bca938da00be950468ba1c25a5cb50e9f6d8aa13c8cd21f24ba909402775b262ac76d374d",
			  tx: {
				accessList: [{
					address: "0x0000000000000000000000000000000000001337",
					storageKeys: [...]
				}],
				chainId: "0x796f6c6f763378",
				gas: "0x62d4",
				gasPrice: "0x3b9aca00",
				hash: "0xbbd570a3c6acc9bb7da0d5c0322fe4ea2a300db80226f7df4fef39b2d6649eec",
				input: "0x",
				nonce: "0x0",
				r: "0x294ac94077b35057971e6b4b06dfdf55a6fbed819133a6c1d31e187f1bca938d",
				s: "0xbe950468ba1c25a5cb50e9f6d8aa13c8cd21f24ba909402775b262ac76d374d",
				to: "0xdf0a88b2b68c673713a8ec826003676f272e3573",
				type: "0x1",
				v: "0x0",
				value: "0x1"
			  }
			}
	*/
	tx := eth.Transaction{
		Type:     eth.MustQuantity("0x1"),
		Gas:      eth.QuantityFromInt64(0x62d4),
		GasPrice: eth.QuantityFromInt64(0x3b9aca00),
		Input:    eth.Data("0x"),
		Nonce:    eth.QuantityFromInt64(0),
		To:       eth.MustAddress("0xdf0a88b2b68c673713a8ec826003676f272e3573"),
		Value:    eth.QuantityFromInt64(0x1),
		AccessList: &eth.AccessList{
			eth.AccessListEntry{
				Address: "0x0000000000000000000000000000000000001337",
				StorageKeys: []eth.Data32{
					*eth.MustData32("0x0000000000000000000000000000000000000000000000000000000000000000"),
				},
			},
		},
	}

	expectedUnsigned := "0x01f86587796f6c6f76337880843b9aca008262d494df0a88b2b68c673713a8ec826003676f272e35730180f838f7940000000000000000000000000000000000001337e1a00000000000000000000000000000000000000000000000000000000000000000808080"
	unsigned, err := tx.RawRepresentation(chainId)
	require.NoError(t, err)
	require.Equal(t, expectedUnsigned, unsigned.String())

	signed, err := tx.Sign("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19", chainId)
	require.NoError(t, err)

	expectedSigned := "0x01f8a587796f6c6f76337880843b9aca008262d494df0a88b2b68c673713a8ec826003676f272e35730180f838f7940000000000000000000000000000000000001337e1a0000000000000000000000000000000000000000000000000000000000000000080a0294ac94077b35057971e6b4b06dfdf55a6fbed819133a6c1d31e187f1bca938da00be950468ba1c25a5cb50e9f6d8aa13c8cd21f24ba909402775b262ac76d374d"
	require.Equal(t, expectedSigned, signed.String())
	require.Equal(t, "0xbbd570a3c6acc9bb7da0d5c0322fe4ea2a300db80226f7df4fef39b2d6649eec", tx.Hash.String())
	require.Equal(t, "0x294ac94077b35057971e6b4b06dfdf55a6fbed819133a6c1d31e187f1bca938d", tx.R.String())
	require.Equal(t, "0xbe950468ba1c25a5cb50e9f6d8aa13c8cd21f24ba909402775b262ac76d374d", tx.S.String())
	require.Equal(t, "0x0", tx.V.String())
}
