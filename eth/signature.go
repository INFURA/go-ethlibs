package eth

import (
	"encoding/hex"

	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"

	secp256k1 "github.com/btcsuite/btcd/btcec"
)

type Signature struct {
	r       Quantity
	s       Quantity
	v       Quantity
	chainId Quantity
}

// NewEIP2718Signature creates a new Signature from discrete chainId, R, S, and V values.
func NewEIP2718Signature(chainId Quantity, r Quantity, s Quantity, v Quantity) (*Signature, error) {
	if chainId.Int64() == 0 {
		return nil, errors.New("chainId is required for EIP-2718 style signatures")
	}
	if vi := v.Int64(); vi > 1 || vi < 0 {
		return nil, errors.New("v must be 0x0 or 0x1 for EIP-2718 style signatures")
	}

	sig := Signature{
		r:       r,
		s:       s,
		v:       v,
		chainId: chainId,
	}
	return &sig, nil
}

// NewEIP155Signature creates a new Signature from EIP-155 packed R,S,V values
func NewEIP155Signature(r Quantity, s Quantity, v Quantity) (*Signature, error) {
	sig := Signature{
		r: r,
		s: s,
		v: v,
	}
	// Unpack the passed in v into "standard" v and chainId
	vi := v.Int64()
	switch {
	case vi == 27 || vi == 28:
		sig.v = QuantityFromInt64(vi - 27)
		sig.chainId = QuantityFromInt64(0)
		return &sig, nil
	case vi == 0 || vi == 1:
		sig.v = QuantityFromInt64(vi)
		sig.chainId = QuantityFromInt64(0)
		return &sig, nil
	case vi >= 35:
		// Pull out chainId and recoveryV from EIP-155 packed V
		_chainId := (vi - 35) / 2
		_v := (vi - 27) - ((_chainId * 2) + 8)
		sig.chainId = QuantityFromInt64(_chainId)
		sig.v = QuantityFromInt64(_v)
		return &sig, nil
	default:
		return nil, errors.New("unexpected EIP-155 V value")
	}
}

// ECSign returns the signature values for a given message hash for the given chainId using the bytes of given
// private key.  Primarily used to sign transactions before submitting them with eth_sendRawTransaction.
func ECSign(h *Hash, privKeyBytes []byte, chainId Quantity) (*Signature, error) {
	// code for this method inspired by https://github.com/ethereumjs/ethereumjs-util/blob/master/src/signature.ts#L15
	priv, pub := secp256k1.PrivKeyFromBytes(secp256k1.S256(), privKeyBytes)
	addr, err := pubKeyBytesToAddress(pub.SerializeUncompressed())
	if err != nil {
		return nil, errors.Wrap(err, "could not convert key to ethereum address")
	}

	rawsig, err := priv.Sign(h.Bytes())
	if err != nil {
		return nil, errors.Wrap(err, "signing failed")
	}

	r := QuantityFromBigInt(rawsig.R)
	s := QuantityFromBigInt(rawsig.S)

	// Unfortunately the ECDSA package we are using doesn't return the recovery V value
	// so the only recourse is to try both values 0 or 1 and see which one produces a valid
	// value.
	v := QuantityFromInt64(1)
	sender, err := ECRecover(h, &r, &s, &v)
	if err != nil || sender.String() != addr.String() {
		// ok try the other recovery value
		v = QuantityFromInt64(0)
		sender, err = ECRecover(h, &r, &s, &v)
		if err != nil {
			return nil, errors.Wrap(err, "recovery failed")
		}
	}

	if sender.String() != addr.String() {
		return nil, errors.New("signature mismatch")
	}

	return &Signature{r, s, v, chainId}, nil
}

// EIP155Values returns the expected R,S, and V values for an EIP-155 Signature.  Namely, the V value includes the
// EIP-155 encoded chain id.
func (s *Signature) EIP155Values() (R Quantity, S Quantity, V Quantity) {
	if s.chainId.Int64() == 0 {
		return s.r, s.s, QuantityFromInt64(s.v.Int64() + 27)
	} else {
		return s.r, s.s, QuantityFromInt64(s.v.Int64() + (s.chainId.Int64()*2 + 35))
	}
}

// EIP2718Values returns the expected R, S, and V values for EIP-2718 signatures.  Namely, the V value is simply the 0x0
// or 0x1 parity bit.
func (s *Signature) EIP2718Values() (R Quantity, S Quantity, V Quantity) {
	return s.r, s.s, s.v
}

func (s *Signature) Recover(hash *Hash) (*Address, error) {
	return ECRecover(hash, &s.r, &s.s, &s.v)
}

func (s *Signature) ChainId() (*Quantity, error) {
	if s.chainId.Int64() == 0 {
		return nil, errors.New("chainId was not provided")
	}
	chainId := s.chainId
	return &chainId, nil
}

// ECRecover returns the sending address, given a message digest and R, S, V values.
// Primarily used to recover the sender of eth.Transaction objects.
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
	return pubKeyBytesToAddress(pubKey)
}

// pubKeyBytesToAddress converts the uncompressed bytes of a secp256k1.PublicKey into an
// Ethereum address.
func pubKeyBytesToAddress(uncompressed []byte) (*Address, error) {
	if len(uncompressed) == 0 || uncompressed[0] != 0x4 {
		return nil, errors.New("invalid public key recovered")
	}

	// We'll strip off the 0x04 and hash the rest ...
	remainder := uncompressed[1:]
	hash := sha3.NewLegacyKeccak256()
	hash.Write(remainder)
	sum := hash.Sum(nil)
	digest := hex.EncodeToString(sum)
	keyHash := "0x" + digest

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
