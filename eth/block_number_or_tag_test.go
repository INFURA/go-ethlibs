package eth_test

import (
	"encoding/json"
	"github.com/INFURA/go-ethlibs/eth"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBlockNumberOrTag_MarshalJSON(t *testing.T) {

	{
		tag := eth.MustBlockNumberOrTag("latest")
		b, err := json.Marshal(&tag)
		require.NoError(t, err)
		require.Equal(t, []byte(`"latest"`), b)
	}

	{
		num := eth.MustBlockNumberOrTag("0x7654321")
		b, err := json.Marshal(&num)
		require.NoError(t, err)
		require.Equal(t, []byte(`"0x7654321"`), b)
	}

	{
		s := struct {
			Tag eth.BlockNumberOrTag `json:"tag"`
		}{
			Tag: *eth.MustBlockNumberOrTag("latest"),
		}

		b, err := json.Marshal(s)
		require.NoError(t, err)
		require.Equal(t, []byte(`{"tag":"latest"}`), b)
	}
}
