package eth_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/INFURA/ethereum-interaction/pkg/eth"
	"github.com/INFURA/ethereum-interaction/pkg/jsonrpc"
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
}
