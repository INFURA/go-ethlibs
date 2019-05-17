package rlp

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func DecodeString(input string) ([]string, error) {
	if !strings.HasPrefix(input, "0x") {
		return nil, errors.New("invalid hex input")
	}

	output := make([]string, 0)
	remainder := input[2:]
	for {
		if len(remainder) == 0 {
			return output, nil
		}

		b := remainder[0:2]
		prefix, err := strconv.ParseUint(b, 16, 8)
		if err != nil {
			return nil, errors.Wrap(err, "could not decode prefix")
		}

		switch {
		case 0x00 <= prefix && prefix <= 0x7f:
			// single byte value, append it
			output = append(output, "0x"+remainder[0:2])
			remainder = remainder[2:]
			continue
		case 0x80 <= prefix && prefix <= 0xb7:
			// short string
			size := (prefix - 0x80) * 2
			if size > uint64(len(remainder)) {
				return nil, errors.New("insufficient remaining input for short string")
			}
			remainder = remainder[2:]
			output = append(output, "0x"+remainder[0:size])
			remainder = remainder[size:]
			continue
		case 0xb8 <= prefix && prefix <= 0xbf:
			// long string
			sizeSize := int((prefix - 0xb7) * 2)
			if sizeSize > len(remainder) {
				return nil, errors.New("insufficient remaining input for size of long string")
			}
			remainder = remainder[2:]

			size, err := strconv.ParseUint(remainder[0:sizeSize], 16, 64)
			if err != nil {
				return nil, errors.Wrap(err, "could not decode long string size")
			}
			size *= 2
			remainder = remainder[sizeSize:]

			if size > uint64(len(remainder)) {
				return nil, errors.New("insufficient remaining input for long string")
			}

			output = append(output, "0x"+remainder[0:size])
			remainder = remainder[size:]
			continue
		case 0xc0 <= prefix && prefix <= 0xf7:
			// short list
			size := (prefix - 0xc0) * 2
			// copy the list as is
			if size > uint64(len(remainder)) {
				return nil, errors.New("insufficient remaining input for short list")
			}
			remainder = remainder[2:]
			output = append(output, "0x"+remainder[0:size])
			remainder = remainder[size:]
			continue
		case 0xf8 <= prefix && prefix <= 0xff:
			// long list
			sizeSize := int((prefix - 0xf7) * 2)
			if sizeSize > len(remainder) {
				return nil, errors.New("insufficient remaining input for size of long list")
			}
			remainder = remainder[2:]
			size, err := strconv.ParseInt(remainder[0:sizeSize], 16, 64)
			if err != nil {
				return nil, errors.Wrap(err, "could not decode long list size")
			}
			size *= 2
			remainder = remainder[sizeSize:]

			// copy the list as is
			if size > int64(len(remainder)) {
				return nil, errors.New("insufficient remaining input for short list")
			}
			output = append(output, "0x"+remainder[0:size])
			remainder = remainder[size:]
			continue
		}
	}

}
