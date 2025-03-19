package eth_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	`github.com/justinwongcn/go-ethlibs/eth`
)

func TestNewAccessListFromRLP(t *testing.T) {
	{
		// One address with one key
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

	{
		// Multiple addresses, multiple keys
		src := eth.AccessList{
			eth.AccessListEntry{
				Address: "0x0000000000000000000000000000000000001337",
				StorageKeys: []eth.Data32{
					*eth.MustData32("0x0000000000000000000000000000000000000000000000000000000000000000"),
				},
			},
			eth.AccessListEntry{
				Address: "0x0000000000000000000000000000000000004444",
				StorageKeys: []eth.Data32{
					*eth.MustData32("0x0000000000000000000000000000000000000000000000000000000000001234"),
					*eth.MustData32("0x0000000000000000000000000000000000000000000000000000000000005678"),
				},
			},
		}

		asRLP := src.RLP()
		encoded, err := asRLP.Encode()
		require.NoError(t, err)
		require.Equal(t, "0xf893f7940000000000000000000000000000000000001337e1a00000000000000000000000000000000000000000000000000000000000000000f859940000000000000000000000000000000000004444f842a00000000000000000000000000000000000000000000000000000000000001234a00000000000000000000000000000000000000000000000000000000000005678", encoded)

		decoded, err := eth.NewAccessListFromRLP(asRLP)
		require.NoError(t, err)

		require.Equal(t, src, decoded)
	}

	{
		// One address no keys
		src := eth.AccessList{
			eth.AccessListEntry{
				Address:     "0x0000000000000000000000000000000000001337",
				StorageKeys: []eth.Data32{},
			},
		}

		asRLP := src.RLP()
		encoded, err := asRLP.Encode()
		require.NoError(t, err)
		require.Equal(t, "0xd7d6940000000000000000000000000000000000001337c0", encoded)

		decoded, err := eth.NewAccessListFromRLP(asRLP)
		require.NoError(t, err)

		require.Equal(t, src, decoded)
	}

	{
		// multiple addresses, no keys
		src := eth.AccessList{
			eth.AccessListEntry{
				Address:     "0x0000000000000000000000000000000000001337",
				StorageKeys: []eth.Data32{},
			},
			eth.AccessListEntry{
				Address:     "0x0000000000000000000000000000000000004444",
				StorageKeys: []eth.Data32{},
			},
		}

		asRLP := src.RLP()
		encoded, err := asRLP.Encode()
		require.NoError(t, err)
		require.Equal(t, "0xeed6940000000000000000000000000000000000001337c0d6940000000000000000000000000000000000004444c0", encoded)

		decoded, err := eth.NewAccessListFromRLP(asRLP)
		require.NoError(t, err)

		require.Equal(t, src, decoded)
	}

	{
		// empty list
		src := eth.AccessList{}

		asRLP := src.RLP()
		encoded, err := asRLP.Encode()
		require.NoError(t, err)
		require.Equal(t, "0xc0", encoded)

		decoded, err := eth.NewAccessListFromRLP(asRLP)
		require.NoError(t, err)

		require.Equal(t, src, decoded)
	}
}
