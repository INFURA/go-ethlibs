package eth_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/INFURA/go-ethlibs/eth"
)

func TestTransactionReceipts(t *testing.T) {
	raw := `{
    "blockHash": "0xa37f46c4692db33012c105a27b9e4c582e822ed60a54667875fb92def52fd75a",
    "blockNumber": "0x72991c",
    "contractAddress": null,
    "cumulativeGasUsed": "0x7650c2",
    "from": "0x9e44b7d42125b7bb4e809406ed5e1079ff500969",
    "gasUsed": "0x5630",
    "logs": [
      {
        "address": "0x21ab6c9fac80c59d401b37cb43f81ea9dde7fe34",
        "blockHash": "0xa37f46c4692db33012c105a27b9e4c582e822ed60a54667875fb92def52fd75a",
        "blockNumber": "0x72991c",
        "data": "0x000000000000000000000000000000000000000000000000000000070560c8c0",
        "logIndex": "0xb7",
        "removed": false,
        "topics": [
          "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
          "0x0000000000000000000000009e44b7d42125b7bb4e809406ed5e1079ff500969",
          "0x000000000000000000000000fe5854255eb1eb921525fa856a3947ed2412a1d7"
        ],
        "transactionHash": "0x9d2fb08850a9b38173044ae6a61974fde4eacca504e399ffd9d5c8af567113cc",
        "transactionIndex": "0x8a"
      }
    ],
    "logsBloom": "0x00000001000000000000000000000000000000000000000000008000000000000000000000010000000000000000000000400000000000000000000000000008000000000000000000000008000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200000000002000000000000000000000000000000000000000000000000000000000000000008000000000000000000000010000000000000000000000000000000",
    "status": "0x1",
    "to": "0x21ab6c9fac80c59d401b37cb43f81ea9dde7fe34",
    "transactionHash": "0x9d2fb08850a9b38173044ae6a61974fde4eacca504e399ffd9d5c8af567113cc",
    "transactionIndex": "0x8a"
  }`

	receipt := eth.TransactionReceipt{}
	err := json.Unmarshal([]byte(raw), &receipt)
	require.NoError(t, err, "unmarshal must succeed")

	require.Equal(t, uint64(0x72991c), receipt.BlockNumber.UInt64())
	require.Equal(t, "0xa37f46c4692db33012c105a27b9e4c582e822ed60a54667875fb92def52fd75a", receipt.BlockHash.String())
	require.Nil(t, receipt.ContractAddress)
	require.Equal(t, uint64(0x7650c2), receipt.CumulativeGasUsed.UInt64())
	require.Equal(t, uint64(0x5630), receipt.GasUsed.UInt64())
	require.Equal(t, *eth.MustAddress("0x9e44b7d42125b7bb4e809406ed5e1079ff500969"), receipt.From)
	require.Equal(t, eth.MustAddress("0x21ab6c9fac80c59d401b37cb43f81ea9dde7fe34"), receipt.To)
	require.Equal(t, uint64(1), receipt.Status.UInt64())
	require.Equal(t, *eth.MustHash("0x9d2fb08850a9b38173044ae6a61974fde4eacca504e399ffd9d5c8af567113cc"), receipt.TransactionHash)
	require.Equal(t, uint64(0x8a), receipt.TransactionIndex.UInt64())

	require.Equal(t, *eth.MustTopic("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"), receipt.Logs[0].Topics[0])
	require.Equal(t, *eth.MustData256("0x00000001000000000000000000000000000000000000000000008000000000000000000000010000000000000000000000400000000000000000000000000008000000000000000000000008000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200000000002000000000000000000000000000000000000000000000000000000000000000008000000000000000000000010000000000000000000000000000000"), receipt.LogsBloom)
	require.Nil(t, receipt.Type)
	require.Equal(t, eth.TransactionTypeLegacy, receipt.TransactionType())

	// double check that we can back back to JSON as well
	b, err := json.Marshal(&receipt)
	require.NoError(t, err, "marshal must succeed")
	require.JSONEq(t, raw, string(b))

	// Let double check that contract creation receipts work too
	creation := `{
    "blockHash": "0xaacadbbc77f8962c0f2749ca12145ddfd09857c4ef4d6caa507d0afda7c200f5",
    "blockNumber": "0x578049",
    "contractAddress": "0x21ab6c9fac80c59d401b37cb43f81ea9dde7fe34",
    "cumulativeGasUsed": "0x4fac0d",
    "from": "0x4ea0d7225e384582d6ea31e34260bf7ac0c1127f",
    "gasUsed": "0xe85d5",
    "logs": [],
    "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
    "status": "0x1",
    "to": null,
    "transactionHash": "0x5e07f2daaa7fad0d59a19ba2bce54d1b58be25f5c8dd684e1f27e59c3d5f92c6",
    "transactionIndex": "0x6a"
  }`

	{
		receipt := eth.TransactionReceipt{}
		err := json.Unmarshal([]byte(creation), &receipt)
		require.NoError(t, err, "unmarshal must succeed")

		require.Equal(t, eth.MustAddress("0x21ab6c9fac80c59d401b37cb43f81ea9dde7fe34"), receipt.ContractAddress)
		require.Nil(t, receipt.To)
		require.Nil(t, receipt.Type)
		require.Equal(t, eth.TransactionTypeLegacy, receipt.TransactionType())
	}

	// And a pre-byzantine one while we're here
	old := `{
    "blockHash": "0xcd6d29f6b644e82252823053c2e051bab2461f24d3d32b7bb2e5391452f2386e",
    "blockNumber": "0x7a122",
    "contractAddress": null,
    "cumulativeGasUsed": "0x1a7a1",
    "from": "0x119058dc2c577e9c4ba6914678aa9db565300ffe",
    "gasUsed": "0x723c",
    "logs": [
      {
        "address": "0x46a9a148d617138cb5c0346de289c030856bb716",
        "blockHash": "0xcd6d29f6b644e82252823053c2e051bab2461f24d3d32b7bb2e5391452f2386e",
        "blockNumber": "0x7a122",
        "data": "0x000000000000000000000000119058dc2c577e9c4ba6914678aa9db565300ffe000000000000000000000000000000000000000000000a968163f0a57b400000",
        "logIndex": "0x1",
        "removed": false,
        "topics": [
          "0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c"
        ],
        "transactionHash": "0x45215aa0da9b7597d233d96b6f7c4ac311edaba77a99ecc6471c59663554914f",
        "transactionIndex": "0x1"
      }
    ],
    "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000008000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000400000000000000000",
    "root": "0x57bd5108d8f0b8bad735ab77e2a47b80c166dcf5059b2960e0118b40562c7cf2",
    "to": "0x46a9a148d617138cb5c0346de289c030856bb716",
    "transactionHash": "0x45215aa0da9b7597d233d96b6f7c4ac311edaba77a99ecc6471c59663554914f",
    "transactionIndex": "0x1"
  }`

	{
		receipt := eth.TransactionReceipt{}
		err := json.Unmarshal([]byte(old), &receipt)
		require.NoError(t, err, "unmarshal must succeed")

		require.Equal(t, eth.MustData32("0x57bd5108d8f0b8bad735ab77e2a47b80c166dcf5059b2960e0118b40562c7cf2"), receipt.Root)
		require.Nil(t, receipt.Status)
		require.Nil(t, receipt.Type)
		require.Equal(t, eth.TransactionTypeLegacy, receipt.TransactionType())
	}

	// EIP-2718 receipts
	legacy := `{
    "type": "0x0",
    "blockHash": "0xa37f46c4692db33012c105a27b9e4c582e822ed60a54667875fb92def52fd75a",
    "blockNumber": "0x72991c",
    "contractAddress": null,
    "cumulativeGasUsed": "0x7650c2",
    "from": "0x9e44b7d42125b7bb4e809406ed5e1079ff500969",
    "gasUsed": "0x5630",
    "logs": [
      {
        "address": "0x21ab6c9fac80c59d401b37cb43f81ea9dde7fe34",
        "blockHash": "0xa37f46c4692db33012c105a27b9e4c582e822ed60a54667875fb92def52fd75a",
        "blockNumber": "0x72991c",
        "data": "0x000000000000000000000000000000000000000000000000000000070560c8c0",
        "logIndex": "0xb7",
        "removed": false,
        "topics": [
          "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
          "0x0000000000000000000000009e44b7d42125b7bb4e809406ed5e1079ff500969",
          "0x000000000000000000000000fe5854255eb1eb921525fa856a3947ed2412a1d7"
        ],
        "transactionHash": "0x9d2fb08850a9b38173044ae6a61974fde4eacca504e399ffd9d5c8af567113cc",
        "transactionIndex": "0x8a"
      }
    ],
    "logsBloom": "0x00000001000000000000000000000000000000000000000000008000000000000000000000010000000000000000000000400000000000000000000000000008000000000000000000000008000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200000000002000000000000000000000000000000000000000000000000000000000000000008000000000000000000000010000000000000000000000000000000",
    "status": "0x1",
    "to": "0x21ab6c9fac80c59d401b37cb43f81ea9dde7fe34",
    "transactionHash": "0x9d2fb08850a9b38173044ae6a61974fde4eacca504e399ffd9d5c8af567113cc",
    "transactionIndex": "0x8a"
  }`

	{
		receipt := eth.TransactionReceipt{}
		err := json.Unmarshal([]byte(legacy), &receipt)
		require.NoError(t, err, "unmarshal must succeed")

		require.NotNil(t, receipt.Type)
		require.Equal(t, eth.TransactionTypeLegacy, receipt.Type.Int64())
		require.Equal(t, eth.TransactionTypeLegacy, receipt.TransactionType())
	}

	eip2930 := `{
		"type": "0x1",
		"blockHash": "0xc6b65d9a251257942744ba1f250df218c2db4c1ec91d54d505034af5029f5edc",
		"blockNumber": "0x45",
		"contractAddress": null,
		"cumulativeGasUsed": "0x62d4",
		"from": "0x8a8eafb1cf62bfbeb1741769dae1a9dd47996192",
		"gasUsed": "0x62d4",
		"logs": [],
		"logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		"status": "0x1",
		"to": "0x8a8eafb1cf62bfbeb1741769dae1a9dd47996192",
		"transactionHash": "0x5cb00d928abf074cb81fc5e54dd49ef541afa7fc014b8a53fb8c29f3ecb5cadb",
		"transactionIndex": "0x0"
	  }`

	{
		receipt := eth.TransactionReceipt{}
		err := json.Unmarshal([]byte(eip2930), &receipt)
		require.NoError(t, err, "unmarshal must succeed")

		require.NotNil(t, receipt.Type)
		require.Equal(t, eth.TransactionTypeAccessList, receipt.Type.Int64())
		require.Equal(t, eth.TransactionTypeAccessList, receipt.TransactionType())
	}

}

func TestTransactionReceipt_4844(t *testing.T) {
	// curl https://rpc.dencun-devnet-8.ethpandaops.io/ -H 'Content-Type: application/json' -d '{"method":"eth_getTransactionReceipt","params":["0x5ceec39b631763ae0b45a8fb55c373f38b8fab308336ca1dc90ecd2b3cf06d00"],"id":1,"jsonrpc":"2.0"}'
	raw := `{
		"blobGasPrice": "0x1",
		"blobGasUsed": "0x20000",
		"blockHash": "0xfc2715ff196e23ae613ed6f837abd9035329a720a1f4e8dce3b0694c867ba052",
		"blockNumber": "0x2a1cb",
		"contractAddress": null,
		"cumulativeGasUsed": "0x5208",
		"effectiveGasPrice": "0x1d1a94a201c",
		"from": "0xad01b55d7c3448b8899862eb335fbb17075d8de2",
		"gasUsed": "0x5208",
		"logs": [],
		"logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		"status": "0x1",
		"to": "0x000000000000000000000000000000000000f1c1",
		"transactionHash": "0x5ceec39b631763ae0b45a8fb55c373f38b8fab308336ca1dc90ecd2b3cf06d00",
		"transactionIndex": "0x0",
		"type": "0x3"
	  }`

	receipt := eth.TransactionReceipt{}
	err := json.Unmarshal([]byte(raw), &receipt)
	require.NoError(t, err, "unmarshal must succeed")

	require.NotNil(t, receipt.Type)
	require.Equal(t, "0x3", receipt.Type.String())
	require.Equal(t, "0x1", receipt.BlobGasPrice.String())
	require.Equal(t, "0x20000", receipt.BlobGasUsed.String())

	// convert back to JSON and compare
	b, err := json.Marshal(&receipt)
	require.NoError(t, err)
	require.JSONEq(t, raw, string(b))
}

func TestTransactionReceipt_RLP(t *testing.T) {
	t.Run("Sepolia 1559 Receipt", func(t *testing.T) {
		payload := `{"blockHash":"0x59d58395078eb687813eeade09f4ccd3a40084e607c3b0e0b987794c12be48cc","blockNumber":"0xd0275","contractAddress":null,"cumulativeGasUsed":"0xaebb","effectiveGasPrice":"0x59682f07","from":"0x1b57edab586cbdabd4d914869ae8bb78dbc05571","gasUsed":"0xaebb","logs":[{"address":"0x830bf80a3839b300291915e7c67b70d90823ffed","blockHash":"0x59d58395078eb687813eeade09f4ccd3a40084e607c3b0e0b987794c12be48cc","blockNumber":"0xd0275","data":"0x0000000000000000000000000000000000000000000000000de0b6b3a7640000","logIndex":"0x0","removed":false,"topics":["0xe1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c","0x0000000000000000000000001b57edab586cbdabd4d914869ae8bb78dbc05571"],"transactionHash":"0x30296f5f32972c7c3b39963cfd91073000cb882c294adc2dcf0ac9ca34d67bd2","transactionIndex":"0x0"}],"logsBloom":"0x00800000000000000000000000000000000000000000100000020000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000800000000000000400000000000000000","status":"0x1","to":"0x830bf80a3839b300291915e7c67b70d90823ffed","transactionHash":"0x30296f5f32972c7c3b39963cfd91073000cb882c294adc2dcf0ac9ca34d67bd2","transactionIndex":"0x0","type":"0x2"}`
		receipt := eth.TransactionReceipt{}
		err := json.Unmarshal([]byte(payload), &receipt)
		require.NoError(t, err)

		require.Equal(t, eth.TransactionTypeDynamicFee, receipt.TransactionType())

		raw, err := receipt.RawRepresentation()
		require.NoError(t, err)
		expectedRawRepresentation := `0x02f901850182aebbb9010000800000000000000000000000000000000000000000100000020000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000800000000000000400000000000000000f87cf87a94830bf80a3839b300291915e7c67b70d90823ffedf842a0e1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109ca00000000000000000000000001b57edab586cbdabd4d914869ae8bb78dbc05571a00000000000000000000000000000000000000000000000000de0b6b3a7640000`

		require.Equal(t, expectedRawRepresentation, raw.String())

		/*
			// TODO: we don't have trie hashing support needed to verify the receipts root
			t.Run("receipts root", func(t *testing.T) {

				// The above transaction was the only one in Sepolia block 0x59d58395078eb687813eeade09f4ccd3a40084e607c3b0e0b987794c12be48cc with the following receipts root
				expectedReceiptRoot := `0x7800894d3a17b7f4ce8f17f96740e13696982605164eb4465bdd8a313d0953a5`
			})
		*/
	})
	t.Run("Mainnet Pre-Byzantium Receipt", func(t *testing.T) {
		payload := `{"blockHash":"0x11d68b50f327f5ebac40b9487cacf4b6c6fb8ddabd852bbddac16dfc2d4ca6a7","blockNumber":"0x22dd9c","contractAddress":null,"cumulativeGasUsed":"0x47b760","effectiveGasPrice":"0x5d21dba00","from":"0x33daedabab9085bd1a94460a652e7ffff592dfe3","gasUsed":"0x47b760","logs":[],"logsBloom":"0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000","root":"0xc2b35ef55562a517f53c5a8eab65255b6c68950a7b980d2d0e970190c3528228","to":"0x33ccbd4db4372b8e829343d781e2f949223ff4b1","transactionHash":"0xf94b37f170c234eacc203d683808bba6f671f12d8e87c71d246d4aee03deb579","transactionIndex":"0x0","type":"0x0"}`
		receipt := eth.TransactionReceipt{}
		err := json.Unmarshal([]byte(payload), &receipt)
		require.NoError(t, err)

		require.Equal(t, eth.TransactionTypeLegacy, receipt.TransactionType())

		raw, err := receipt.RawRepresentation()
		require.NoError(t, err)
		expectedRawRepresentation := `0xf90129a0c2b35ef55562a517f53c5a8eab65255b6c68950a7b980d2d0e970190c35282288347b760b9010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c0`

		require.Equal(t, expectedRawRepresentation, raw.String())

		/*
			// TODO: we don't have trie hashing support needed to verify the receipts root
			t.Run("receipts root", func(t *testing.T) {

				// The above transaction was the only one in Mainnet block 0x11d68b50f327f5ebac40b9487cacf4b6c6fb8ddabd852bbddac16dfc2d4ca6a7 with the following receipts root
				expectedReceiptRoot := `0x45be296e6c6b0215eeee992909e19e659c224452ae04cb00f7662e65fce81ce1`
			})
		*/
	})
}
