package eth_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/INFURA/go-ethlibs/eth"
	"github.com/INFURA/go-ethlibs/jsonrpc"
)

func TestQuantityFromUInt64(t *testing.T) {
	q := eth.QuantityFromUInt64(uint64(0x1234))

	b, err := json.Marshal(&q)
	require.NoError(t, err)
	require.Equal(t, []byte(`"0x1234"`), b)

	params := jsonrpc.MustParams(&q, true)
	b, err = json.Marshal(&params)
	require.NoError(t, err)

	require.Equal(t, []byte(`["0x1234",true]`), b)

	zero := eth.Quantity{}
	b, err = json.Marshal(&zero)
	require.NoError(t, err)
	require.Equal(t, []byte(`"0x0"`), b)

	fromBig := eth.QuantityFromBigInt(big.NewInt(0x4567))
	b, err = json.Marshal(&fromBig)
	require.NoError(t, err)

	require.Equal(t, []byte(`"0x4567"`), b)

	fromUI64 := eth.QuantityFromUInt64(0xABCD)
	b, err = json.Marshal(&fromUI64)
	require.NoError(t, err)

	require.Equal(t, []byte(`"0xabcd"`), b)
}
