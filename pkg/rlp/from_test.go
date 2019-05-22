package rlp_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/INFURA/ethereum-interaction/pkg/rlp"
)

func TestFrom(t *testing.T) {
	{
		// This is a list (actually a raw transaction)
		input := "0xf86d820144843b9aca0082520894b78777860637d56543da23312c7865024833f7d188016345785d8a0000802ba0e2539a5d9f056d7095bd19d6b77b850910eeafb71534ebd45159915fab202e91a007484420f3968697974413fc55d1142dc76285d30b1b9231ccb71ed1e720faae"

		decoded, err := rlp.From(input)
		require.NoError(t, err)
		require.Equal(t, 9, len(decoded.List))

		payload := decoded.List
		require.Equal(t, "0x0144", payload[0].String, "Nonce")
		require.Equal(t, "0x3b9aca00", payload[1].String, "Price")
		require.Equal(t, "0x5208", payload[2].String, "GasLimit")
		require.Equal(t, "0xb78777860637d56543da23312c7865024833f7d1", payload[3].String, "To")
		require.Equal(t, "0x016345785d8a0000", payload[4].String, "Amount")
		require.Equal(t, "0x", payload[5].String, "Data")
		require.Equal(t, "0x2b", payload[6].String, "V")
		require.Equal(t, "0xe2539a5d9f056d7095bd19d6b77b850910eeafb71534ebd45159915fab202e91", payload[7].String, "R")
		require.Equal(t, "0x07484420f3968697974413fc55d1142dc76285d30b1b9231ccb71ed1e720faae", payload[8].String, "S")

		encoded, err := decoded.Encode()
		require.NoError(t, err)
		require.Equal(t, input, encoded)
	}

	// Some examples from https://github.com/ethereum/tests/blob/develop/RLPTests/rlptest.json
	//  Copyright 2014 Ethereum Foundation - MIT Licensed
	{
		// multilist: [ "zw", [ 4 ], 1 ]
		input := "0xc6827a77c10401"
		decoded, err := rlp.From(input)
		require.NoError(t, err)
		require.Equal(t, 3, len(decoded.List))
		require.Equal(t, "0x7a77", decoded.List[0].String)
		require.Equal(t, "0x04", decoded.List[1].List[0].String)
		require.Equal(t, "0x01", decoded.List[2].String)
	}

	{
		// listsoflists: [ [ [], [] ], [] ]
		input := "0xc4c2c0c0c0"
		decoded, err := rlp.From(input)
		require.NoError(t, err)
		require.Equal(t, 2, len(decoded.List))
		require.Equal(t, 2, len(decoded.List[0].List))
		require.Equal(t, 0, len(decoded.List[0].List[0].List))
		require.Equal(t, 0, len(decoded.List[0].List[1].List))
		require.Equal(t, 0, len(decoded.List[1].List))
	}

	{
		// shortListMax1: [ "asdf", "qwer", "zxcv", "asdf","qwer", "zxcv", "asdf", "qwer", "zxcv", "asdf", "qwer"]
		input := "0xf784617364668471776572847a78637684617364668471776572847a78637684617364668471776572847a78637684617364668471776572"
		decoded, err := rlp.From(input)
		require.NoError(t, err)
		require.Equal(t, 11, len(decoded.List))
		require.Equal(t, "0x61736466", decoded.List[0].String)
		require.Equal(t, "0x71776572", decoded.List[1].String)
		require.Equal(t, "0x7a786376", decoded.List[2].String)
		require.Equal(t, decoded.List[0].String, decoded.List[3].String)
		require.Equal(t, decoded.List[1].String, decoded.List[4].String)
		require.Equal(t, decoded.List[2].String, decoded.List[5].String)
		require.Equal(t, decoded.List[3].String, decoded.List[6].String)
		require.Equal(t, decoded.List[4].String, decoded.List[7].String)
		require.Equal(t, decoded.List[5].String, decoded.List[8].String)
		require.Equal(t, decoded.List[6].String, decoded.List[9].String)
		require.Equal(t, decoded.List[7].String, decoded.List[10].String)
	}

	{
		// bigint: #115792089237316195423570985008687907853269984665640564039457584007913129639936
		input := "0xa1010000000000000000000000000000000000000000000000000000000000000000"
		decoded, err := rlp.From(input)
		require.NoError(t, err)
		require.Equal(t, "0x010000000000000000000000000000000000000000000000000000000000000000", decoded.String)
	}
}
