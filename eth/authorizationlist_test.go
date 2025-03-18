package eth_test

import (
	"testing"

	"github.com/INFURA/go-ethlibs/eth"
	"github.com/stretchr/testify/require"
)

func TestAuthorizationList_FromRLP(t *testing.T) {
	src := eth.AuthorizationList{
		eth.SetCodeAuthorization{
			ChainID: eth.MustQuantity("0x1"),
			Address: *eth.MustAddress("0x000000000000000000000000000000000000aaaa"),
			Nonce:   eth.QuantityFromInt64(0x1),
			V:       eth.QuantityFromInt64(0x1),
			R:       *eth.MustQuantity("0xf7e3e597fc097e71ed6c26b14b25e5395bc8510d58b9136af439e12715f2d721"),
			S:       *eth.MustQuantity("0x6cf7c3d7939bfdb784373effc0ebb0bd7549691a513f395e3cdabf8602724987"),
		},
		eth.SetCodeAuthorization{
			ChainID: eth.MustQuantity("0x0"),
			Address: *eth.MustAddress("0x000000000000000000000000000000000000bbbb"),
			Nonce:   eth.QuantityFromInt64(0x0),
			V:       eth.QuantityFromInt64(0x1),
			R:       *eth.MustQuantity("0x5011890f198f0356a887b0779bde5afa1ed04e6acb1e3f37f8f18c7b6f521b98"),
			S:       *eth.MustQuantity("0x56c3fa3456b103f3ef4a0acb4b647b9cab9ec4bc68fbcdf1e10b49fb2bcbcf61"),
		},
	}

	// encode
	asRLP := src.RLP()
	encoded, err := asRLP.Encode()
	require.NoError(t, err)
	expected := "0xf8b8f85a0194000000000000000000000000000000000000aaaa0101a0f7e3e597fc097e71ed6c26b14b25e5395bc8510d58b9136af439e12715f2d721a06cf7c3d7939bfdb784373effc0ebb0bd7549691a513f395e3cdabf8602724987f85a8094000000000000000000000000000000000000bbbb8001a05011890f198f0356a887b0779bde5afa1ed04e6acb1e3f37f8f18c7b6f521b98a056c3fa3456b103f3ef4a0acb4b647b9cab9ec4bc68fbcdf1e10b49fb2bcbcf61"
	require.Equal(t, expected, encoded, "wrong encoding")

	// get back the original list
	authorizationList, err := eth.NewAuthorizationListFromRLP(asRLP)
	require.NoError(t, err)
	require.Equal(t, len(src), len(authorizationList), "authorization lists have different lengths")

	for i := range src {
		srcAuth := src[i]
		auth := authorizationList[i]

		require.Equal(t, srcAuth.ChainID, auth.ChainID, "authorization lists not equal")
		require.Equal(t, srcAuth.Address, auth.Address, "authorization address not equal")
		require.Equal(t, srcAuth.Nonce.String(), auth.Nonce.String(), "authorization nonce not equal")
		require.Equal(t, srcAuth.V.String(), auth.V.String(), "authorization V not equal")
		require.Equal(t, srcAuth.R.String(), auth.R.String(), "authorization R not equal")
		require.Equal(t, srcAuth.S.String(), auth.S.String(), "authorization S not equal")
	}
}
