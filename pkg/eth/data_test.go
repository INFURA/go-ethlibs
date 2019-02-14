package eth_test

import (
	"github.com/INFURA/eth/pkg/eth"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestData(t *testing.T) {

	require.Equal(t, *eth.MustData("0x"), eth.Data("0x"))
	require.Equal(t, *eth.MustData8("0x0011223344556677"), eth.Data8("0x0011223344556677"))
	require.Equal(t, *eth.MustTopic("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"), eth.Topic("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"))

	var err error
	_, err = eth.NewData256("0x")
	require.Error(t, err)

	_, err = eth.NewData256("0x00")
	require.Error(t, err)
}
