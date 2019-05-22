package rlp

import (
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/pkg/errors"
)

func (v Value) Encode() (string, error) {
	if v.String != "" {
		if !strings.HasPrefix(v.String, "0x") {
			return "", errors.New("invalid string value before encoding")
		}

		b, err := hex.DecodeString(v.String[2:])
		if err != nil {
			return "", errors.Wrap(err, "could not decode string value")
		}

		switch {
		case len(b) == 1 && b[0] <= 0x7f:
			// then the string is it's own encoding
			return v.String, nil
		case len(b) < 56:
			return "0x" + asHex(uint64(0x80+len(b))) + v.String[2:], nil
		default:
			size := asHex(uint64(len(b)))
			sizeSize := uint64(len(size) / 2)
			return "0x" + asHex(0xb7+sizeSize) + size + v.String[2:], nil
		}
	}

	count := len(v.List)
	if count == 0 {
		// return the empty list
		return "0xc0", nil
	}

	data := make([]string, len(v.List))
	for i, item := range v.List {
		encoded, err := item.Encode()
		if err != nil {
			return "", errors.Wrap(err, "could not encode child item")
		}

		// Discard the 0x prefix
		data[i] = encoded[2:]
	}

	body := strings.Join(data, "")
	bodySize := uint64(len(body) / 2)
	if bodySize < 56 {
		// 0xc0 + bodySize
		return "0x" + asHex(bodySize+0xc0) + body, nil
	} else {
		bodySizeHex := asHex(bodySize)
		bodySizeSize := uint64(len(bodySizeHex) / 2)
		return "0x" + asHex(bodySizeSize+0xf7) + bodySizeHex + body, nil
	}
}

func asHex(i uint64) string {
	bn := big.NewInt(0).SetUint64(i)
	b := bn.Bytes()
	h := hex.EncodeToString(b)
	return h
}
