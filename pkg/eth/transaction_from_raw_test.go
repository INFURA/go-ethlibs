package eth_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/INFURA/ethereum-interaction/pkg/eth"
)

func TestTransaction_FromRaw(t *testing.T) {
	input := "0xf86c258502540be40083035b609482e041e84074fc5f5947d4d27e3c44f824b7a1a187b1a2bc2ec500008078a04a7db627266fa9a4116e3f6b33f5d245db40983234eb356261f36808909d2848a0166fa098a2ce3bda87af6000ed0083e3bf7cc31c6686b670bd85cbc6da2d6e85"
	tx := eth.Transaction{}
	err := tx.FromRaw(input)
	require.NoError(t, err)
	require.Equal(t, "0x58e5a0fc7fbc849eddc100d44e86276168a8c7baaa5604e44ba6f5eb8ba1b7eb", tx.Hash.String())
	require.Equal(t, eth.MustAddress("0x6bc84f6a0fabbd7102be338c048fe0ae54948c2e").String(), tx.From.String())
}
