package rlp_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ConsenSys/go-ethlibs/rlp"
)

func fuzz(s string) error {
	if decoded, err := rlp.From(s); err != nil {
		return nil
	} else {
		if _, err := decoded.Encode(); err != nil {
			return err
		}
		return nil
	}
}

// TestFuzzCrashers uses data found by go-fuzz to crash older versions of this library
func TestFuzzCrashers(t *testing.T) {
	inputs := []string{
		"0xc8BF00000000000000",
		"0x8300?000",
		"0xff-018001600070000",
		"0xff7000001604000000",
		"0xf70000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000840000000",
		"0xf09500000000000000000000000000000000000000000000900000000000000000000000000000000000c7000000000000",
		"0xd0c3fa000000000000000000000000000000000000000000",
		"0xf3000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000084000000",
		"0x",
		"0xc48300000",
		"0x0",
		"0x000",
		"0",
		"",
		"0xxx",
		"0xww",
	}

	for _, input := range inputs {
		require.NoError(t, fuzz(input))
	}
}
