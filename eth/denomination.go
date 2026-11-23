package eth

import (
	"math/big"
)

// Ether denomination units, expressed in wei.
//
// These are shared values. Callers must not mutate them; copy first with
// new(big.Int).Set(unit) if you need a value you can change.
//
// Reference: https://ethereum.org/developers/docs/intro-to-ether/#denominations.
var (
	// Wei is the atomic unit: 1 wei, 10⁻¹⁸ ETH.
	Wei = big.NewInt(1)
	// KWei (babbage) is 10³ wei, 10⁻¹⁵ ETH.
	KWei = new(big.Int).Exp(big.NewInt(10), big.NewInt(3), nil)
	// MWei (lovelace) is 10⁶ wei, 10⁻¹² ETH.
	MWei = new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil)
	// GWei (shannon) is 10⁹ wei, 10⁻⁹ ETH.
	GWei = new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil)
	// MicroEther (szabo) is 10¹² wei, 10⁻⁶ ETH.
	MicroEther = new(big.Int).Exp(big.NewInt(10), big.NewInt(12), nil)
	// MilliEther (finney) is 10¹⁵ wei, 10⁻³ ETH.
	MilliEther = new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil)
	// Ether is 10¹⁸ wei, 1 ETH.
	Ether = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
)
