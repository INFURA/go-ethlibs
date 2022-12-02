package eth_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ConsenSys/go-ethlibs/eth"
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

	{
		// invalid addresses should be invalid
		invalid, err := eth.NewAddress("0xinvalid")
		require.Error(t, err, "invalid address must not parse")
		require.Nil(t, invalid, "invalid address must not be returned")

		var addr eth.Address
		err = json.Unmarshal([]byte(`"0xinvalid"`), &addr)
		require.Error(t, err, "invalid address must not parse")
	}
}

type TestAddressNested struct {
	Value   eth.Address  `json:"value"`
	Pointer *eth.Address `json:"pointer"`
}

type TestAddressStruct struct {
	Value         eth.Address        `json:"value"`
	Pointer       *eth.Address       `json:"pointer"`
	Nested        TestAddressNested  `json:"nested"`
	NestedPointer *TestAddressNested `json:"nestedPointer"`
}

func TestAddress_MarshalJSON_Embedded(t *testing.T) {
	addr := "0xEB3Dbe1E1fa0Fefb4F85A692E65e11ff3Eb4F41F"

	t.Run("eth.Log", func(t *testing.T) {
		log := eth.Log{
			Address: *eth.MustAddress(addr),
			Data:    *eth.MustData("0x1234"),
		}

		b, err := json.Marshal(log)
		require.NoError(t, err)
		require.Equal(
			t,
			`{"removed":false,"logIndex":null,"transactionIndex":null,"transactionHash":null,"blockHash":null,"blockNumber":null,"address":"0xeb3dbe1e1fa0fefb4f85a692e65e11ff3eb4f41f","data":"0x1234","topics":null}`,
			string(b),
		)

		b2, err := json.Marshal(&log)
		require.NoError(t, err)
		require.Equal(t, b, b2)
	})

	t.Run("Embedded", func(t *testing.T) {
		s := struct {
			Address eth.Address `json:"address"`
		}{
			Address: *eth.MustAddress(addr),
		}

		b, err := json.Marshal(s)
		require.NoError(t, err)
		require.Equal(t, `{"address":"0xeb3dbe1e1fa0fefb4f85a692e65e11ff3eb4f41f"}`, string(b))

		b2, err := json.Marshal(&s)
		require.NoError(t, err)
		require.Equal(t, b, b2)
	})

	t.Run("Nested", func(t *testing.T) {
		s := TestAddressStruct{
			Value:   *eth.MustAddress(addr),
			Pointer: eth.MustAddress(addr),
			Nested: TestAddressNested{
				Value:   *eth.MustAddress(addr),
				Pointer: eth.MustAddress(addr),
			},
			NestedPointer: &TestAddressNested{
				Value:   *eth.MustAddress(addr),
				Pointer: eth.MustAddress(addr),
			},
		}

		b, err := json.Marshal(s)
		require.NoError(t, err)

		require.JSONEq(t, `{
			"value":"0xeb3dbe1e1fa0fefb4f85a692e65e11ff3eb4f41f", 
			"pointer":"0xeb3dbe1e1fa0fefb4f85a692e65e11ff3eb4f41f", 
			"nested": {
				"value":"0xeb3dbe1e1fa0fefb4f85a692e65e11ff3eb4f41f", 
				"pointer":"0xeb3dbe1e1fa0fefb4f85a692e65e11ff3eb4f41f"
			},
			"nestedPointer": {
				"value":"0xeb3dbe1e1fa0fefb4f85a692e65e11ff3eb4f41f", 
				"pointer":"0xeb3dbe1e1fa0fefb4f85a692e65e11ff3eb4f41f"
			}
		}`, string(b))

		b2, err := json.Marshal(&s)
		require.NoError(t, err)
		require.Equal(t, b, b2)
	})
}
