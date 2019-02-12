package eth_test

import (
	"github.com/INFURA/eth/pkg/eth"
	"github.com/hashicorp/packer/common/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMainnetBlock(t *testing.T) {

	partial := `{
    "difficulty": "0xbfabcdbd93dda",
    "extraData": "0x737061726b706f6f6c2d636e2d6e6f64652d3132",
    "gasLimit": "0x79f39e",
    "gasUsed": "0x79ccd3",
    "hash": "0xb3b20624f8f0f86eb50dd04688409e5cea4bd02d700bf6e79e9384d47d6a5a35",
    "logsBloom": "0x4848112002a2020aaa0812180045840210020005281600c80104264300080008000491220144461026015300100000128005018401002090a824a4150015410020140400d808440106689b29d0280b1005200007480ca950b15b010908814e01911000054202a020b05880b914642a0000300003010044044082075290283516be82504082003008c4d8d14462a8800c2990c88002a030140180036c220205201860402001014040180002006860810ec0a1100a14144148408118608200060461821802c081000042d0810104a8004510020211c088200420822a082040e10104c00d010064004c122692020c408a1aa2348020445403814002c800888208b1",
    "miner": "0x5a0b54d5dc17e0aadc383d2db43b0a0d3e029c4c",
    "mixHash": "0x3d1fdd16f15aeab72e7db1013b9f034ee33641d92f71c0736beab4e67d34c7a7",
    "nonce": "0x4db7a1c01d8a8072",
    "number": "0x5bad55",
    "parentHash": "0x61a8ad530a8a43e3583f8ec163f773ad370329b2375d66433eb82f005e1d6202",
    "receiptsRoot": "0x5eced534b3d84d3d732ddbc714f5fd51d98a941b28182b6efe6df3a0fe90004b",
    "sha3Uncles": "0x8a562e7634774d3e3a36698ac4915e37fc84a2cd0044cb84fa5d80263d2af4f6",
    "size": "0x41c7",
    "stateRoot": "0xf5208fffa2ba5a3f3a2f64ebd5ca3d098978bedd75f335f56b705d8715ee2305",
    "timestamp": "0x5b541449",
    "totalDifficulty": "0x12ac11391a2f3872fcd",
    "transactions": [
      "0x8784d99762bccd03b2086eabccee0d77f14d05463281e121a62abfebcf0d2d5f",
      "0x311be6a9b58748717ac0f70eb801d29973661aaf1365960d159e4ec4f4aa2d7f",
      "0xe42b0256058b7cad8a14b136a0364acda0b4c36f5b02dea7e69bfd82cef252a2"
    ],
    "transactionsRoot": "0xf98631e290e88f58a46b7032f025969039aa9b5696498efc76baf436fa69b262",
    "uncles": [
      "0x824cce7c7c2ec6874b9fa9a9a898eb5f27cbaf3991dfa81084c3af60d1db618c"
    ]
  }`

	var block eth.Block

	err := json.Unmarshal([]byte(partial), &block)
	require.NoError(t, err, "mainnet partial block should deserialize")

	require.Equal(t, 3, len(block.Transactions))
	require.Equal(t, "0x5bad55", block.Number.String())
	require.Equal(t, uint64(0x79f39e), block.GasLimit.UInt64())

	full := `{
    "difficulty": "0x742a575f662",
    "extraData": "0xd783010302844765746887676f312e352e31856c696e7578",
    "gasLimit": "0x2fefd8",
    "gasUsed": "0x5208",
    "hash": "0x648509915efa19b169ccab758492c7525b8498747678b894befd9ff78ad05519",
    "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
    "miner": "0x2a65aca4d5fc5b5c859090a6c34d164135398226",
    "mixHash": "0x47e7eab7d034cf4b8b1501ebfc98edf715ee62f56283bf1a22a5423990600dff",
    "nonce": "0xeacef1c5a2ca3a49",
    "number": "0x99999",
    "parentHash": "0xffa241fbb914038a429c90daeeb54885f31e431d05b12fe87de8007853a1f278",
    "receiptsRoot": "0xb46f767bd3f69c0d7830eae6717f77560ee2ace0ea701d9e95fd41eb39a619ab",
    "sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
    "size": "0x290",
    "stateRoot": "0x93e74cf453c3327075b7e252deeb2d115cf2fdb204ba89806cebbd32afdedaa8",
    "timestamp": "0x565eafba",
    "totalDifficulty": "0x336f973a0249a1e9",
    "transactions": [
      {
        "blockHash": "0x648509915efa19b169ccab758492c7525b8498747678b894befd9ff78ad05519",
        "blockNumber": "0x99999",
        "from": "0x4bb96091ee9d802ed039c4d1a5f6216f90f81b01",
        "gas": "0xa028",
        "gasPrice": "0xba43b7400",
        "hash": "0xb4c724bf1f01a5371c513389d5758d531b729f15c8c6af8f74a100585d2cf33f",
        "input": "0x",
        "nonce": "0x461e",
        "r": "0xd5ee485b95d5992a4ca7d210ff28d540aea3f4031ce39203298ae266bcdb3485",
        "s": "0x71ecb17bdbbae8c57681649a95e8c7e22b90adac2e19c314de3b74ecfb5f8ce1",
        "to": "0x86d3856ad0105b9d4199936c1fd203664ba325dc",
        "transactionIndex": "0x0",
        "v": "0x1b",
        "value": "0x44b1eec6162f0000"
      }
    ],
    "transactionsRoot": "0x237e46a0a93850f7979546c717ffccce6715a6b2cb0bdb0d59a9c559a0d74f07",
    "uncles": []
  }`

	err = json.Unmarshal([]byte(full), &block)
	require.NoError(t, err, "mainnet full block should deserialize")

	require.Equal(t, 1, len(block.Transactions))
	require.Equal(t, "0x99999", block.Number.String())
	require.Equal(t, *block.Hash, block.Transactions[0].BlockHash)
	require.Equal(t, eth.Data("0xd783010302844765746887676f312e352e31856c696e7578"), block.ExtraData)
	require.Equal(t, true, block.Transactions[0].Populated)
	require.Equal(t, int64(0), block.Transactions[0].Index.Int64())
	require.Equal(t, eth.Data("0x"), block.Transactions[0].Input)
}
