package rlp

import (
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func (v *Value) Encode() (string, error) {
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
			return "0x" + strconv.FormatUint(uint64(0x80+len(b)), 16) + v.String[2:], nil
		default:
			size := strconv.FormatUint(uint64(len(b)), 16)
			sizeSize := uint64(len(size) / 2)
			return "0x" + strconv.FormatUint(sizeSize, 16) + size + v.String[2:], nil
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
	bodySize := int64(len(body) / 2)
	if bodySize < 56 {
		// 0xc0 + bodySize
		return "0x" + strconv.FormatInt(bodySize+0xc0, 16) + body, nil
	} else {
		bodySizeHex := strconv.FormatInt(bodySize, 16)
		bodySizeSize := int64(len(bodySizeHex) / 2)
		return "0x" + strconv.FormatInt(bodySizeSize+0xf7, 16) + bodySizeHex + body, nil
	}
}
