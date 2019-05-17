package rlp_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/INFURA/ethereum-interaction/pkg/rlp"
)

func TestDecodeString(t *testing.T) {
	// This is a list (actually a raw transaction)
	input := "0xf86d820144843b9aca0082520894b78777860637d56543da23312c7865024833f7d188016345785d8a0000802ba0e2539a5d9f056d7095bd19d6b77b850910eeafb71534ebd45159915fab202e91a007484420f3968697974413fc55d1142dc76285d30b1b9231ccb71ed1e720faae"

	// Should decode as a single list member
	decoded, err := rlp.DecodeString(input)
	require.NoError(t, err)
	require.Equal(t, 1, len(decoded))

	// Should decode to 9 fields (Nonce, Price, GasLimit, To, Amount, Data, V, R,S)
	contents := decoded[0]
	payload, err := rlp.DecodeString(contents)
	require.NoError(t, err)
	require.Equal(t, 9, len(payload))

	require.Equal(t, "0x0144", payload[0], "Nonce")
	require.Equal(t, "0x3b9aca00", payload[1], "Price")
	require.Equal(t, "0x5208", payload[2], "GasLimit")
	require.Equal(t, "0xb78777860637d56543da23312c7865024833f7d1", payload[3], "To")
	require.Equal(t, "0x016345785d8a0000", payload[4], "Amount")
	require.Equal(t, "0x", payload[5], "Data")
	require.Equal(t, "0x2b", payload[6], "V")
	require.Equal(t, "0xe2539a5d9f056d7095bd19d6b77b850910eeafb71534ebd45159915fab202e91", payload[7], "R")
	require.Equal(t, "0x07484420f3968697974413fc55d1142dc76285d30b1b9231ccb71ed1e720faae", payload[8], "S")
}
