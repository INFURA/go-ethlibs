package rlp

import (
	"encoding/hex"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

func (v Value) Hash() (string, error) {
	encoded, err := v.Encode()
	if err != nil {
		return "", errors.Wrap(err, "could not encode RLP value")
	}

	input := strings.Replace(strings.ToLower(encoded), "0x", "", 1)
	b, err := hex.DecodeString(input)
	if err != nil {
		return "", errors.Wrap(err, "could not convert encoded to bytes")
	}

	hash := sha3.NewLegacyKeccak256()
	hash.Write(b)
	sum := hash.Sum(nil)
	digest := hex.EncodeToString(sum)

	return "0x" + digest, nil
}
