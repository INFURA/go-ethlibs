package rlp_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ConsenSys/go-ethlibs/rlp"
)

func TestValue_Encode(t *testing.T) {
	// From: https://github.com/ethereum/wiki/wiki/RLP#examples
	{
		// The string "dog" = [ 0x83, 'd', 'o', 'g' ]
		// dog = 64 6f 67
		encoded, err := rlp.Value{String: "0x646f67"}.Encode()
		require.NoError(t, err)
		require.Equal(t, "0x83646f67", encoded)
	}

	{
		// The list [ "cat", "dog" ] = [ 0xc8, 0x83, 'c', 'a', 't', 0x83, 'd', 'o', 'g' ]
		// dog = 64 6f 67 cat = 63 61 74
		encoded, err := rlp.Value{
			List: []rlp.Value{
				{String: "0x636174"},
				{String: "0x646f67"},
			},
		}.Encode()
		require.NoError(t, err)
		require.Equal(t, "0xc88363617483646f67", encoded)
	}

	{
		// The empty string ('null') = [ 0x80 ]
		encoded, err := rlp.Value{String: "0x"}.Encode()
		require.NoError(t, err)
		require.Equal(t, "0x80", encoded)
	}

	{
		// The empty list = [ 0xc0 ]
		encoded, err := rlp.Value{}.Encode()
		require.NoError(t, err)
		require.Equal(t, "0xc0", encoded)
	}

	{
		// The encoded integer 0 ('\x00') = [ 0x00 ]
		encoded, err := rlp.Value{String: "0x00"}.Encode()
		require.NoError(t, err)
		require.Equal(t, "0x00", encoded)
	}

	{
		// The encoded integer 15 ('\x0f') = [ 0x0f ]
		encoded, err := rlp.Value{String: "0x0f"}.Encode()
		require.NoError(t, err)
		require.Equal(t, "0x0f", encoded)
	}

	{
		// The encoded integer 1024 ('\x04\x00') = [ 0x82, 0x04, 0x00 ]
		encoded, err := rlp.Value{String: "0x0400"}.Encode()
		require.NoError(t, err)
		require.Equal(t, "0x820400", encoded)
	}
}
