package eth_test

import (
	"encoding/json"
	"github.com/INFURA/eth/pkg/eth"
	"github.com/INFURA/eth/pkg/jsonrpc"
	"github.com/stretchr/testify/require"
	"testing"
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
}
