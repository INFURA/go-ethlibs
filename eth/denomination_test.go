package eth_test

import (
	"math/big"
	"testing"

	"github.com/INFURA/go-ethlibs/eth"
	"github.com/stretchr/testify/require"
)

func TestDenominationConversion(t *testing.T) {
	tests := []struct {
		name          string
		unit          *big.Int
		multipleOfWei string
		valueInETH    string
	}{
		{"wei", eth.Wei, "1", "1/1000000000000000000"},
		{"kwei", eth.KWei, "1000", "1/1000000000000000"},
		{"mwei", eth.MWei, "1000000", "1/1000000000000"},
		{"gwei", eth.GWei, "1000000000", "1/1000000000"},
		{"microether", eth.MicroEther, "1000000000000", "1/1000000"},
		{"milliether", eth.MilliEther, "1000000000000000", "1/1000"},
		{"ether", eth.Ether, "1000000000000000000", "1/1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			multipleOfWei := tt.unit.String()
			require.Equal(t, tt.multipleOfWei, multipleOfWei)

			valueInETH := new(big.Rat).SetFrac(tt.unit, eth.Ether).String()
			require.Equal(t, tt.valueInETH, valueInETH)
		})
	}
}
