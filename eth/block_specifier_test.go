package eth_test

import (
	"testing"

	"github.com/INFURA/go-ethlibs/eth"
	"github.com/stretchr/testify/require"
)

func TestBlockSpecifierMarshalUnmarshal(t *testing.T) {
	{
		var spec eth.BlockSpecifier
		spec.UnmarshalJSON([]byte(`"0x0"`))
		require.Equal(t, "0x0", spec.Number.String())
		require.Nil(t, spec.Hash)
		require.Nil(t, spec.Tag)
		require.Nil(t, spec.RequireCanonical)

		m, err := spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`{"blockNumber":"0x0"}`), m)

		spec.Raw = true
		m, err = spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`"0x0"`), m)
	}

	{
		var spec eth.BlockSpecifier
		spec.UnmarshalJSON([]byte(`{ "blockNumber": "0x0" }`))
		require.Equal(t, "0x0", spec.Number.String())
		require.Nil(t, spec.Hash)
		require.Nil(t, spec.Tag)
		require.Nil(t, spec.RequireCanonical)

		m, err := spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`{"blockNumber":"0x0"}`), m)

		spec.Raw = true
		m, err = spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`"0x0"`), m)
	}

	{
		var spec eth.BlockSpecifier
		spec.UnmarshalJSON([]byte(`
			{
				"blockHash": "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"
			}
		`))
		require.Equal(t, "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3", spec.Hash.String())
		require.Equal(t, false, *spec.RequireCanonical)
		require.Nil(t, spec.Number)
		require.Nil(t, spec.Tag)

		m, err := spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`{"blockHash":"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3","requireCanonical":false}`), m)

		spec.Raw = true
		m, err = spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`{"blockHash":"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3","requireCanonical":false}`), m)
	}

	{
		var spec eth.BlockSpecifier
		spec.UnmarshalJSON([]byte(`
			{ 
				"blockHash": "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3",
				"requireCanonical": false 
			}
		`))
		require.Equal(t, "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3", spec.Hash.String())
		require.Equal(t, false, *spec.RequireCanonical)
		require.Nil(t, spec.Number)
		require.Nil(t, spec.Tag)

		m, err := spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`{"blockHash":"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3","requireCanonical":false}`), m)

		spec.Raw = true
		m, err = spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`{"blockHash":"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3","requireCanonical":false}`), m)
	}

	{
		var spec eth.BlockSpecifier
		spec.UnmarshalJSON([]byte(`
			{ 
				"blockHash": "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3",
				"requireCanonical": true 
			}
		`))
		require.Equal(t, "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3", spec.Hash.String())
		require.Equal(t, true, *spec.RequireCanonical)
		require.Nil(t, spec.Number)
		require.Nil(t, spec.Tag)

		m, err := spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`{"blockHash":"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3","requireCanonical":true}`), m)

		spec.Raw = true
		m, err = spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`{"blockHash":"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3","requireCanonical":true}`), m)
	}
}

// { "blockHash": "0x<non-existent-block-hash>" } -> raise block-not-found error
// { "blockHash": "0x<non-existent-block-hash>", "requireCanonical": false } -> raise block-not-found error
// { "blockHash": "0x<non-existent-block-hash>", "requireCanonical": true } -> raise block-not-found error
// { "blockHash": "0x<non-canonical-block-hash>" } -> return storage at given address in specified block
// { "blockHash": "0x<non-canonical-block-hash>", "requireCanonical": false } -> return storage at given address in specified block
// { "blockHash": "0x<non-canonical-block-hash>", "requireCanonical": true } -> raise block-not-canonical error
