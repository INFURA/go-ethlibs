package rlp

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func from(input string) (*Value, string, error) {
	remainder := input
	b := remainder[0:2]
	prefix, err := strconv.ParseUint(b, 16, 8)
	if err != nil {
		return nil, "", errors.Wrap(err, "could not decode prefix")
	}

	switch {
	case 0x00 <= prefix && prefix <= 0x7f:
		// single byte value, append it
		s := remainder[0:2]
		remainder = remainder[2:]
		return &Value{String: "0x" + s}, remainder, nil
	case 0x80 <= prefix && prefix <= 0xb7:
		// short string
		size := (prefix - 0x80) * 2
		if size > uint64(len(remainder)) {
			return nil, "", errors.New("insufficient remaining input for short string")
		}
		remainder = remainder[2:]
		s := remainder[0:size]
		remainder = remainder[size:]
		return &Value{String: "0x" + s}, remainder, nil
	case 0xb8 <= prefix && prefix <= 0xbf:
		// long string
		sizeSize := int((prefix - 0xb7) * 2)
		if sizeSize > len(remainder) {
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
	case 0xc0 <= prefix && prefix <= 0xf7:
		// short list
		size := (prefix - 0xc0) * 2
		// copy the list as is
		if size > uint64(len(remainder)) {
			return nil, "", errors.New("insufficient remaining input for short list")
		}
		remainder = remainder[2:]
		l, err := list(remainder[0:size])
		if err != nil {
			return nil, "", err
		}
		remainder = remainder[size:]
		return &Value{List: l}, remainder, nil
	case 0xf8 <= prefix && prefix <= 0xff:
		// long list
		sizeSize := int((prefix - 0xf7) * 2)
		if sizeSize > len(remainder) {
			return nil, "", errors.New("insufficient remaining input for size of long list")
		}
		remainder = remainder[2:]
		size, err := strconv.ParseInt(remainder[0:sizeSize], 16, 64)
		if err != nil {
			return nil, "", errors.Wrap(err, "could not decode long list size")
		}
		size *= 2
		remainder = remainder[sizeSize:]

		// copy the list as is
		if size > int64(len(remainder)) {
			return nil, "", errors.New("insufficient remaining input for short list")
		}
		l, err := list(remainder[0:size])
		if err != nil {
			return nil, "", err
		}
		remainder = remainder[size:]
		return &Value{List: l}, remainder, nil
	}

	panic("Cannot get here")
}

func list(input string) ([]Value, error) {
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
