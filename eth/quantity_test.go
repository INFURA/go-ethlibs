package eth_test

import (
	"encoding/json"
	"math/big"
	"sync"
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

	require.Equal(t, fromUI64.Big().Int64(), eth.QuantityFromInt64(fromUI64.Int64()).Int64())

	require.Equal(t, int64(1), eth.MustQuantity("0x1").Int64())

	{
		// leading zeroes are acceptable on input
		zeroes, err := eth.NewQuantity("0x01230")
		require.NoError(t, err)
		require.Equal(t, uint64(0x1230), zeroes.UInt64())
	}

	{
		invalid, err := eth.NewQuantity("bad")
		require.Error(t, err)
		require.Nil(t, invalid)
	}

	{
		invalid, err := eth.NewQuantity("0xinvalid")
		require.Error(t, err)
		require.Nil(t, invalid)
	}

	{
		invalid, err := eth.NewQuantity("0x")
		require.Error(t, err)
		require.Nil(t, invalid)
	}

	{
		invalid, err := eth.NewQuantity("0")
		require.Error(t, err)
		require.Nil(t, invalid)
	}
}

func TestQuantity_DeepCopyInto(t *testing.T) {
	val := int64(0x123456)
	src := eth.QuantityFromInt64(val)

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			// This is basically the same pattern as the auto-generated eth.Block.DeepCopy
			// code here: https://github.com/INFURA/go-ethlibs/blob/master/eth/zz_deepcopy_generated.go#L19
			cpy := src
			src.DeepCopyInto(&cpy)

			if cpy.Int64() != val {
				panic("not equal")
			}
		}()
	}
}

func TestQuantity_MarshalJSON(t *testing.T) {
	s := struct {
		Value   eth.Quantity  `json:"value"`
		Pointer *eth.Quantity `json:"pointer"`
	}{
		Value:   eth.QuantityFromInt64(0x1234),
		Pointer: eth.MustQuantity("0x1234"),
	}

	expected := `{"value":"0x1234","pointer":"0x1234"}`

	b, err := json.Marshal(s)
	require.NoError(t, err)

	require.JSONEq(t, expected, string(b))

	b2, err := json.Marshal(&s)
	require.NoError(t, err)
	require.Equal(t, string(b), string(b2))
}
