package eth_test

import (
	"encoding/json"
	"github.com/INFURA/eth/pkg/eth"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestAddressChecksums(t *testing.T) {
	expected := []string{
		"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed",
		"0xfB6916095ca1df60bB79Ce92cE3Ea74c37c5d359",
		"0xdbF03B407c01E7cD3CBea99509d93f8DDDC8C6FB",
		"0xD1220A0cf47c7B9Be7A2E6BA89F429762e7b9aDb",
	}

	for _, expected := range expected {
		require.Equal(t, expected, eth.ToChecksumAddress(strings.ToLower(expected)))
	}
}

func TestAddressJSON(t *testing.T) {
	body := `["0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed",
		"0xfB6916095ca1df60bB79Ce92cE3Ea74c37c5d359",
		"0xdbF03B407c01E7cD3CBea99509d93f8DDDC8C6FB",
		"0xD1220A0cf47c7B9Be7A2E6BA89F429762e7b9aDb"]`

	addresses := make([]eth.Address, 0)
	err := json.Unmarshal([]byte(body), &addresses)
	require.NoError(t, err)

	// create some string arrays to compare to
	checksummed := make([]string, 0)
	err = json.Unmarshal([]byte(body), &checksummed)
	require.NoError(t, err)

	lowered := make([]string, len(checksummed))
	copy(lowered, checksummed)
	for i, s := range checksummed {
		lowered[i] = strings.ToLower(s)
	}

	// The in-memory representation of addresses should match checksummed input
	for i, a := range addresses {
		require.Equal(t, checksummed[i], a.String())
	}

	// the JSON produced by an array of addresses and lower-cased strings should match
	aj, err := json.Marshal(&addresses)
	require.NoError(t, err)

	lj, err := json.Marshal(&lowered)
	require.NoError(t, err)

	require.Equal(t, lj, aj)
}
