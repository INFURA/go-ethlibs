package rlp

import (
	"encoding/hex"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

// Hash returns the keccak256 hash of the encoded RLP value as a hexadecimal string prefixed with 0x
func (v Value) Hash() (string, error) {
	// TODO: Consider operating on the already encoded string vs. encoding inside this function
	// Encode the value back to a hex string
	encoded, err := v.Encode()
	if err != nil {
		return "", errors.Wrap(err, "could not encode RLP value")
	}

	// Convert the string to bytes
	input := strings.Replace(strings.ToLower(encoded), "0x", "", 1)
	b, err := hex.DecodeString(input)
	if err != nil {
		return "", errors.Wrap(err, "could not convert encoded to bytes")
	}

	// And feed the bytes into our hash
	hash := sha3.NewLegacyKeccak256()
	hash.Write(b)
	sum := hash.Sum(nil)

	// and finally return the hash as a 0x prefixed string
	digest := hex.EncodeToString(sum)
	return "0x" + digest, nil
}
