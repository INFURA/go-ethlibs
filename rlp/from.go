package rlp

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// From parses a 0x prefixed hex string into an rlp.Value
func From(input string) (*Value, error) {
	if !strings.HasPrefix(input, "0x") {
		return nil, errors.New("invalid hex input")
	}

	value, remainder, err := from(input[2:])
	if err != nil {
		return nil, err
	}
	if remainder != "" {
		return nil, errors.New("extra data at end")
	}
	return value, nil
}

// from parses the input string and returns an rlp.Value and any remaining unparsed text
func from(input string) (*Value, string, error) {

	// This code was heavily assisted by this series of articles:
	//   https://medium.com/coinmonks/ethereum-under-the-hood-part-3-rlp-decoding-c0c07f5c0714
	// And of course the RLP wiki spec:
	//   https://github.com/ethereum/wiki/wiki/RLP

	if input == "" {
		return &Value{String: "0x"}, "", nil
	}

	for _, r := range input {
		if (r < 'a' || 'f' < r) && (r < 'A' || 'F' < r) && (r < '0' || '9' < r) {
			return nil, "", errors.New("invalid rune in input")
		}
	}

	remainder := input
	b := remainder[0:2]
	p, err := strconv.ParseUint(b, 16, 8)
	if err != nil {
		return nil, "", errors.Wrap(err, "could not decode prefix")
	}

	prefix := byte(p)

	switch {
	// 0x00 - 0x7f - For a single byte whose value is in the [0x00, 0x7f] range, that byte is its own RLP encoding.
	case /*0x00 <= prefix &&*/ prefix <= 0x7f:
		// single byte value, append it
		s := remainder[0:2]
		remainder = remainder[2:]
		return &Value{String: "0x" + s}, remainder, nil

	// 0x80 - 0xb7 - Otherwise, if a string is 0-55 bytes long, the RLP encoding consists of a single byte with value
	//               0x80 plus the length of the string followed by the string. The range of the first byte is thus
	//               [0x80, 0xb7]
	case 0x80 <= prefix && prefix <= 0xb7:
		// short string
		size := uint64((prefix - 0x80) * 2)
		if size > uint64(len(remainder)-2) {
			return nil, "", errors.New("insufficient remaining input for short string")
		}
		remainder = remainder[2:]
		s := remainder[0:size]
		remainder = remainder[size:]
		return &Value{String: "0x" + s}, remainder, nil

	// 0xbb - 0xbf - If a string is more than 55 bytes long, the RLP encoding consists of a single byte with value
	//               0xb7 plus the length in bytes of the length of the string in binary form, followed by the length
	//               of the string, followed by the string. For example, a length-1024 string would be encoded
	//               as \xb9\x04\x00 followed by the string. The range of the first byte is thus [0xb8, 0xbf].
	case 0xb8 <= prefix && prefix <= 0xbf:
		// long string
		sizeSize := int((prefix - 0xb7) * 2)
		if sizeSize > len(remainder)-2 {
			return nil, "", errors.New("insufficient remaining input for size of long string")
		}
		remainder = remainder[2:]

		size, err := strconv.ParseUint(remainder[0:sizeSize], 16, 64)
		if err != nil {
			return nil, "", errors.Wrap(err, "could not decode long string size")
		}
		size *= 2
		remainder = remainder[sizeSize:]

		if size > uint64(len(remainder)) {
			return nil, "", errors.New("insufficient remaining input for long string")
		}

		s := remainder[0:size]
		remainder = remainder[size:]
		return &Value{String: "0x" + s}, remainder, nil

	// 0xc0 - 0xf7 - If the total payload of a list (i.e. the combined length of all its items being RLP encoded) is
	//               0-55 bytes long, the RLP encoding consists of a single byte with value 0xc0 plus the length of the
	//               list followed by the concatenation of the RLP encodings of the items. The range of the first byte
	//               is thus [0xc0, 0xf7]
	case 0xc0 <= prefix && prefix <= 0xf7:
		// short list
		size := uint64((prefix - 0xc0) * 2)
		// copy the list as is
		if size > uint64(len(remainder)-2) {
			return nil, "", errors.New("insufficient remaining input for short list")
		}
		remainder = remainder[2:]
		l, err := parseListItems(remainder[0:size])
		if err != nil {
			return nil, "", err
		}
		remainder = remainder[size:]
		return &Value{List: l}, remainder, nil

	// 0xf8 - 0xff - If the total payload of a list is more than 55 bytes long, the RLP encoding consists of a single
	//               byte with value 0xf7 plus the length in bytes of the length of the payload in binary form,
	//               followed by the length of the payload, followed by the concatenation of the RLP encodings of the
	//               items. The range of the first byte is thus [0xf8, 0xff]
	case 0xf8 <= prefix /*&& prefix <= 0xff*/ :
		// long list
		sizeSize := int((prefix - 0xf7) * 2)
		if sizeSize > len(remainder)-2 {
			return nil, "", errors.New("insufficient remaining input for size of long list")
		}
		remainder = remainder[2:]
		size, err := strconv.ParseInt(remainder[0:sizeSize], 16, 64)
		if err != nil {
			return nil, "", errors.Wrap(err, "could not decode long list size")
		}

		maxsize := int64(^uint64(0) >> 2)
		if size > maxsize || size < 0 {
			return nil, "", errors.New("invalid list size")
		}
		size *= 2
		remainder = remainder[sizeSize:]

		// copy the list as is
		if size > int64(len(remainder)) {
			return nil, "", errors.New("insufficient remaining input for short list")
		}
		l, err := parseListItems(remainder[0:size])
		if err != nil {
			return nil, "", err
		}
		remainder = remainder[size:]
		return &Value{List: l}, remainder, nil
	}

	// The golang compiler should recognize that the above switch is exhaustive but doesn't
	panic("unreachable")
}

// parseListItems breaks an RLP text string into an []rlp.Value slice
func parseListItems(input string) ([]Value, error) {
	l := make([]Value, 0)
	for {
		if input == "" {
			break
		}

		v, remainder, err := from(input)
		if err != nil {
			return nil, err
		}

		l = append(l, *v)
		input = remainder
	}

	return l, nil
}
