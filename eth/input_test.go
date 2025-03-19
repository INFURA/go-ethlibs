package eth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInput_FunctionSelector(t *testing.T) {
	t.Run("empty input is nil", func(t *testing.T) {
		input, err := NewInput("0x")
		require.NoError(t, err)
		require.Nil(t, input.FunctionSelector())
	})

	t.Run("short input is nil", func(t *testing.T) {
		input, err := NewInput("0x1234")
		require.NoError(t, err)
		require.Nil(t, input.FunctionSelector())
	})

	t.Run("exact input returns selector", func(t *testing.T) {
		input, err := NewInput("0x2fbbe334")
		require.NoError(t, err)
		require.Equal(t, input.FunctionSelector(), MustData4("0x2fbbe334"))
	})

	t.Run("input with arg returns selector", func(t *testing.T) {
		input, err := NewInput("0xa41368620000000000000000000000000000000000000000000000000000000000000020")
		require.NoError(t, err)
		require.Equal(t, input.FunctionSelector(), MustData4("0xa4136862"))
	})
}
