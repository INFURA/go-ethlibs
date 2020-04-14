package eth_test

import (
	"encoding/hex"
	"testing"

	secp256k1 "github.com/btcsuite/btcd/btcec"
	"github.com/stretchr/testify/require"

	"github.com/INFURA/go-ethlibs/eth"
)

func TestECSignAndRecover(t *testing.T) {
	key, _ := hex.DecodeString("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19")
	secp256k1.PrivKeyFromBytes(secp256k1.S256(), key)
	digest := eth.MustHash("0x40340296657f4ca5b25addda7b14d31458cbf1efab963e949daef0e84415c5dc")
	chainId := eth.QuantityFromInt64(1)

	sig, err := eth.ECSign(digest, key, chainId)
	require.NoError(t, err)
	require.NotNil(t, sig)
}
