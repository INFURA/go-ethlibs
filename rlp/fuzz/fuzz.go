//go:build gofuzz
// +build gofuzz

package rlp

import "github.com/INFURA/go-ethlibs/rlp"

func Fuzz(data []byte) int {
	s := string(data)
	decoded, err := rlp.From(s)
	if err != nil {
		return 0
	}

	if _, err := decoded.Encode(); err != nil {
		panic(err)
	}

	return 1
}
