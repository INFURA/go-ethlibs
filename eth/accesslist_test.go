package eth_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/INFURA/go-ethlibs/eth"
)

func TestNewAccessListFromRLP(t *testing.T) {
	src := eth.AccessList{
		eth.AccessListEntry{
			Address: "0x0000000000000000000000000000000000001337",
			StorageKeys: []eth.Data32{
				*eth.MustData32("0x0000000000000000000000000000000000000000000000000000000000000000"),
			},
		},
	}

	asRLP := src.RLP()
	encoded, err := asRLP.Encode()
	require.NoError(t, err)
	require.Equal(t, "0xf838f7940000000000000000000000000000000000001337e1a00000000000000000000000000000000000000000000000000000000000000000", encoded)

	decoded, err := eth.NewAccessListFromRLP(asRLP)
	require.NoError(t, err)

	require.Equal(t, src, decoded)
}
