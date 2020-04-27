package eth_test

import (
	"testing"

	"github.com/INFURA/go-ethlibs/eth"
	"github.com/stretchr/testify/require"
)

func TestNewBlockSpecifier(t *testing.T) {
	{
		data := "0x0"
		spec, err := eth.NewBlockSpecifier(data)
		require.Nil(t, err)
		require.Equal(t, "0x0", spec.Number.String())
		require.Nil(t, spec.Hash)
		require.Nil(t, spec.Tag)
		require.Equal(t, false, spec.RequireCanonical)
	}

	{
		data := "latest"
		spec, err := eth.NewBlockSpecifier(data)
		require.Nil(t, err)
		require.Nil(t, spec.Number)
		require.Nil(t, spec.Hash)
		require.Equal(t, "latest", spec.Tag.String())
		require.Equal(t, false, spec.RequireCanonical)
	}

	{
		data := "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"
		spec, err := eth.NewBlockSpecifier(data)
		require.Nil(t, err)
		require.Nil(t, spec.Number)
		require.Equal(t, "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3", spec.Hash.String())
		require.Nil(t, spec.Tag)
		require.Equal(t, false, spec.RequireCanonical)
	}

	{
		data := map[string]interface{}{
			"blockNumber": "0x0",
		}
		spec, err := eth.NewBlockSpecifier(data)
		require.Nil(t, err)
		require.Equal(t, "0x0", spec.Number.String())
		require.Nil(t, spec.Hash)
		require.Nil(t, spec.Tag)
		require.Equal(t, false, spec.RequireCanonical)
	}

	{
		data := map[string]interface{}{
			"blockHash": "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3",
		}
		spec, err := eth.NewBlockSpecifier(data)
		require.Nil(t, err)
		require.Nil(t, spec.Number)
		require.Equal(t, "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3", spec.Hash.String())
		require.Nil(t, spec.Tag)
		require.Equal(t, false, spec.RequireCanonical)
	}

	{
		data := map[string]interface{}{
			"blockHash":        "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3",
			"requireCanonical": false,
		}
		spec, err := eth.NewBlockSpecifier(data)
		require.Nil(t, err)
		require.Nil(t, spec.Number)
		require.Equal(t, "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3", spec.Hash.String())
		require.Nil(t, spec.Tag)
		require.Equal(t, false, spec.RequireCanonical)
	}

	{
		data := map[string]interface{}{
			"blockHash":        "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3",
			"requireCanonical": true,
		}
		spec, err := eth.NewBlockSpecifier(data)
		require.Nil(t, err)
		require.Nil(t, spec.Number)
		require.Equal(t, "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3", spec.Hash.String())
		require.Nil(t, spec.Tag)
		require.Equal(t, true, spec.RequireCanonical)
	}
}

func TestBlockSpecifierMarshalUnmarshal(t *testing.T) {
	{
		var spec eth.BlockSpecifier
		spec.UnmarshalJSON([]byte(`"0x0"`))
		require.Equal(t, "0x0", spec.Number.String())
		require.Nil(t, spec.Hash)
		require.Nil(t, spec.Tag)
		require.Equal(t, false, spec.RequireCanonical)

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
		spec.UnmarshalJSON([]byte(`"latest"`))
		require.Equal(t, "latest", spec.Tag.String())
		require.Nil(t, spec.Number)
		require.Nil(t, spec.Hash)
		require.Equal(t, false, spec.RequireCanonical)

		m, err := spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`"latest"`), m)

		spec.Raw = true
		m, err = spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`"latest"`), m)
	}

	{
		var spec eth.BlockSpecifier
		spec.UnmarshalJSON([]byte(`"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"`))
		require.Nil(t, spec.Number)
		require.Equal(t, "0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3", spec.Hash.String())
		require.Nil(t, spec.Tag)
		require.Equal(t, false, spec.RequireCanonical)

		m, err := spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`{"blockHash":"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3","requireCanonical":false}`), m)

		spec.Raw = true
		m, err = spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"`), m)
	}

	{
		var spec eth.BlockSpecifier
		spec.UnmarshalJSON([]byte(`
			{
				"blockNumber": "0x0"
			}
		`))
		require.Equal(t, "0x0", spec.Number.String())
		require.Nil(t, spec.Hash)
		require.Nil(t, spec.Tag)
		require.Equal(t, false, spec.RequireCanonical)

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
		require.Equal(t, false, spec.RequireCanonical)
		require.Nil(t, spec.Number)
		require.Nil(t, spec.Tag)

		m, err := spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`{"blockHash":"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3","requireCanonical":false}`), m)

		spec.Raw = true
		m, err = spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"`), m)
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
		require.Equal(t, false, spec.RequireCanonical)
		require.Nil(t, spec.Number)
		require.Nil(t, spec.Tag)

		m, err := spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`{"blockHash":"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3","requireCanonical":false}`), m)

		spec.Raw = true
		m, err = spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"`), m)
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
		require.Equal(t, true, spec.RequireCanonical)
		require.Nil(t, spec.Number)
		require.Nil(t, spec.Tag)

		m, err := spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`{"blockHash":"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3","requireCanonical":true}`), m)

		spec.Raw = true
		m, err = spec.MarshalJSON()
		require.Nil(t, err)
		require.Equal(t, []byte(`"0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"`), m)
	}
}
