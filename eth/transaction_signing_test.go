package eth_test

import (
    "log"
    // "crypto/ecdsa"
    "testing"

    "github.com/stretchr/testify/require"

    "github.com/INFURA/go-ethlibs/eth"
)

func TestSignTransaction(t *testing.T) {
    data := []byte("0x")
    //tx, err := eth.NewTransaction(5, 21488430592, 90000, "0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed", 1, data)
    tx, err := eth.NewTransaction(0, 21488430592, 90000, "0xc149Be1bcDFa69a94384b46A1F91350E5f81c1AB", 950000000000000000, data)
    require.NoError(t, err)
    require.Equal(t, tx.Nonce.UInt64(), uint64(0))
    require.Equal(t, tx.GasPrice, eth.QuantityFromInt64(21488430592))
    require.Equal(t, tx.Gas, eth.QuantityFromInt64(90000))
    require.Equal(t, tx.To.String(), "0xc149Be1bcDFa69a94384b46A1F91350E5f81c1AB")
    require.Equal(t, tx.Value, eth.QuantityFromInt64(950000000000000000))

    signed, err := tx.Sign("fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19", 1)
    require.NoError(t, err)
    log.Println("rawTx: ", signed)

    tx2 := eth.Transaction{}
    err = tx2.FromRaw(signed)
    require.NoError(t, err)
    require.Equal(t, tx2.Nonce.UInt64(), uint64(0))
    require.Equal(t, tx2.GasPrice, eth.QuantityFromInt64(21488430592))
    require.Equal(t, tx2.Gas, eth.QuantityFromInt64(90000))
    require.Equal(t, tx2.To.String(), "0xc149Be1bcDFa69a94384b46A1F91350E5f81c1AB")
    require.Equal(t, tx2.Value, eth.QuantityFromInt64(950000000000000000))
}


func TestSignTransaction2(t *testing.T) {
    data := []byte("0x")
    tx, err := eth.NewTransaction(146, 3000000000, 22000, "0x43700db832E9Ac990D36d6279A846608643c904E", 1000000000, data)
    require.NoError(t, err)
    require.Equal(t, tx.Nonce, eth.QuantityFromInt64(146))
    require.Equal(t, tx.GasPrice, eth.QuantityFromInt64(3000000000))
    require.Equal(t, tx.Gas, eth.QuantityFromInt64(22000))
    require.Equal(t, tx.To.String(), "0x43700db832E9Ac990D36d6279A846608643c904E")
    require.Equal(t, tx.Value, eth.QuantityFromInt64(1000000000))

    signed, err := tx.Sign("678174D637194A1ACC1A8C2FB494F62DAC9EB3CBCDFD9B9EF182417F73BA21C8", 1)
    require.NoError(t, err)
    log.Println("rawTx: ", signed)

    tx2 := eth.Transaction{}
    err = tx2.FromRaw(signed)
    require.NoError(t, err)
    require.Equal(t, tx2.Nonce, eth.QuantityFromInt64(146))
    require.Equal(t, tx2.GasPrice, eth.QuantityFromInt64(3000000000))
    require.Equal(t, tx2.Gas, eth.QuantityFromInt64(22000))
    require.Equal(t, tx2.To.String(), "0x43700db832E9Ac990D36d6279A846608643c904E")
    require.Equal(t, tx2.Value, eth.QuantityFromInt64(1000000000))
}

func TestSignTransaction3(t *testing.T) {
    data := []byte("0x")
    tx, err := eth.NewTransaction(145, 3000000000, 22000, "0x43700db832E9Ac990D36d6279A846608643c904E", 1000000000, data)
    require.NoError(t, err)
    require.Equal(t, tx.Nonce, eth.QuantityFromInt64(145))
    require.Equal(t, tx.GasPrice, eth.QuantityFromInt64(3000000000))
    require.Equal(t, tx.Gas, eth.QuantityFromInt64(22000))
    require.Equal(t, tx.To.String(), "0x43700db832E9Ac990D36d6279A846608643c904E")
    require.Equal(t, tx.Value, eth.QuantityFromInt64(1000000000))

    signed, err := tx.Sign("678174D637194A1ACC1A8C2FB494F62DAC9EB3CBCDFD9B9EF182417F73BA21C8", 1)
    require.NoError(t, err)
    log.Println("rawTx: ", signed)

}