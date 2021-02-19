package eth_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/INFURA/go-ethlibs/eth"
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

	_, err = eth.NewData("0xfoodbarr")
	require.Error(t, err)

	_, err = eth.NewData("badf00d")
	require.Error(t, err)

	{
		d, err := eth.NewData("0x")
		require.NoError(t, err)
		require.Equal(t, "0x", d.String())
		require.Equal(t, "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470", d.Hash().String())
	}

	{
		d, err := eth.NewData8("0x1122334455667788")
		require.NoError(t, err)
		require.NotNil(t, d)
		require.Equal(t, "0x1360118a9c9fd897720cf4e26de80683f402dd7c28e000aa98ea51b85c60161c", d.Hash().String())
	}

	{
		d, err := eth.NewData20("0x1122334455667788990011223344556677889900")
		require.NoError(t, err)
		require.NotNil(t, d)
		require.Equal(t, "0x0a2fb1c97af2de8f8ac02909daafec285f8ebc8817cb7dc7c606ea892eece1be", d.Hash().String())
	}

	{
		d, err := eth.NewData32("0x1122334455667788990011223344556677889900112233445566778899001122")
		require.NoError(t, err)
		require.NotNil(t, d)
		require.Equal(t, "0xf88d9246fe5c20db67700433fa1048f8dcd2204cd4ab5c52f36f1d027e51505c", d.Hash().String())

		h, err := eth.NewHash(d.String())
		require.NoError(t, err)
		require.NotNil(t, h)

		tt, err := eth.NewTopic(d.String())
		require.NoError(t, err)
		require.NotNil(t, tt)
	}

	{
		d, err := eth.NewData256("0x" + strings.Repeat("00", 256))
		require.NoError(t, err)
		require.NotNil(t, d)
		require.Equal(t, "0xd397b3b043d87fcd6fad1291ff0bfd16401c274896d8c63a923727f077b8e0b5", d.Hash().String(), d.String())
	}
}
