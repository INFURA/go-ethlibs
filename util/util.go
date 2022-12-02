package util

import (
	"fmt"
	"strconv"
)

const (
	base10    int = 10
	bitSize64 int = 64
)

// ParseUint64 converts a string input to an uint64.
func ParseUint64(data string) (uint64, error) {
	result, err := strconv.ParseUint(data, base10, bitSize64)
	if err != nil {
		return 0, fmt.Errorf("%w", err)
	}

	return result, nil
}

// FormatInt64 converts an int64 input to a string.
func FormatInt64(data int64) string {
	return strconv.FormatInt(data, base10)
}
