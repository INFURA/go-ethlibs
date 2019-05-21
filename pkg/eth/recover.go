package eth

import (
	"encoding/hex"

	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"

	secp256k1 "github.com/btcsuite/btcd/btcec"
)

func ECRecover(h *Hash, r, s, v *Quantity) (*Address, error) {
	// TODO: Currently there's no clean way to do this in pure Go
	// However, there is a PR to the go runtime to allow arbitrary A instead of A=-3 (secp256k1 uses A=0):
	//  https://github.com/golang/go/pull/26873/files
	// So, we could just inline this PR in our repo and use that until the PR is accepted (if ever)

	// recover the public key
	hashBytes, err := hex.DecodeString(h.String()[2:])
	if err != nil {
		return nil, errors.Wrap(err, "could not convert hash to bytes")
	}

	vb := byte(v.Big().Uint64() + 27)
	rb, sb := r.Big().Bytes(), s.Big().Bytes()
	sig := make([]byte, 65)
	sig[0] = vb
	copy(sig[33-len(rb):33], rb)
	copy(sig[65-len(sb):65], sb)

	key, _, err := secp256k1.RecoverCompact(secp256k1.S256(), sig, hashBytes)
	if err != nil {
		return nil, errors.Wrap(err, "could not recover secp256k1 key")
	}

	pubKey := key.SerializeUncompressed()
	if len(pubKey) == 0 || pubKey[0] != 0x4 {
		return nil, errors.New("invalid public key recovered")
	}

	keyHash, err := hash(pubKey[1:])
	if err != nil {
		return nil, errors.Wrap(err, "could not hash public key")
	}

	addrSrc := "0x" + keyHash[26:]
	addr, err := NewAddress(addrSrc)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert hash to address")
	}

	return addr, nil
}

func hash(input []byte) (string, error) {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(input)
	sum := hash.Sum(nil)
	digest := hex.EncodeToString(sum)
	return "0x" + digest, nil
}
