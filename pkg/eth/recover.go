package eth

import (
	"encoding/hex"

	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"

	secp256k1 "github.com/btcsuite/btcd/btcec"
)

func ECRecover(h *Hash, r, s, v *Quantity) (*Address, error) {
	// TODO: Currently there's no clean way to do this in with the Go runtime
	// However, there is a PR to the go runtime to allow arbitrary A instead of A=-3 (secp256k1 uses A=0):
	//   https://github.com/golang/go/pull/26873/files
	// In theory, we could just inline this PR in our repo and use that until the PR is accepted (if ever)

	// For the meantime, we will use btcd's eliptic curve implementation.
	// The code below is based heavily on the secp256k1.recoverAddress implementation at:
	//   https://github.com/ethers-io/ethers.js/blob/34397fa2aaa9187f307881ec10f07dc035dc0854/src.ts/utils/secp256k1.ts#L109
	// with some trial and error to get working with btcd.
	// I also used some of the code changes proposed in:
	//   https://github.com/tendermint/tendermint/pull/3441
	// To determine that btcd is capable of recovering ethereuem addresses.

	// recover the public key
	hashBytes, err := hex.DecodeString(h.String()[2:])
	if err != nil {
		return nil, errors.Wrap(err, "could not convert hash to bytes")
	}

	// NOTE: btcd's secp256k1 expects V at offset 0 NOT offset 64
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

	// Get the UNCOMPRESSED public key, which should always start with a 0x04
	pubKey := key.SerializeUncompressed()
	if len(pubKey) == 0 || pubKey[0] != 0x4 {
		return nil, errors.New("invalid public key recovered")
	}

	// We'll strip off the 0x04 and hash the rest ...
	keyHash, err := hash(pubKey[1:])
	if err != nil {
		return nil, errors.Wrap(err, "could not hash public key")
	}

	// ... and then your Ethereum address is the last 20 bytes of said hash
	// since the hash is already a hex string of 66 characters ( 0x + 32x2 )
	// strip off the first 26 characters and re-add the 0x to get the last 20 bytes
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
