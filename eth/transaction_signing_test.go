package eth_test

import (
   // "crypto/ecdsa"
    "testing"
    "log"

    "github.com/stretchr/testify/require"

    "github.com/INFURA/go-ethlibs/eth"
)

func TestSignTransaction(t *testing.T) {
    data := []byte("0x")
    tx, err := eth.NewTransaction(5, 21488430592, 90000, "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed", 1, data)
    require.NoError(t, err)
    require.Equal(t, tx.Nonce, *eth.MustQuantity("0x5"))
    require.Equal(t, tx.GasPrice, eth.QuantityFromInt64(21488430592))
    require.Equal(t, tx.Gas, eth.QuantityFromInt64(90000))
    require.Equal(t, tx.To.String(), "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed")
    require.Equal(t, tx.Value, eth.QuantityFromInt64(1))

    signed, err := tx.Sign("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19", 1)
    require.NoError(t, err)
    log.Println(signed)
}
