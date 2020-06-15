package eth

import (
	"encoding/hex"

	"golang.org/x/crypto/sha3"
)

// +k8s:deepcopy-gen=false
type Bloom struct {
	value [256]byte
}

func (b *Bloom) Value() Data256 {
	return Data256("0x" + hex.EncodeToString(b.value[:]))
}

func (b *Bloom) AddLog(log Log) {
	b.AddAddress(log.Address)
	for _, topic := range log.Topics {
		b.AddData32(topic)
	}
}

func (b *Bloom) AddAddress(addr Address) {
	b.AddBytes(addr.Bytes())
}

func (b *Bloom) AddData32(data Data32) {
	b.AddBytes(data.Bytes())
}

func (b *Bloom) AddBytes(_bytes []byte) {
	for _, bits := range b.bloomBits(_bytes) {
		b.set(bits)
	}
}

func (b *Bloom) MatchesLog(log Log) bool {
	if !b.MatchesAddress(log.Address) {
		return false
	}
	for _, topic := range log.Topics {
		if !b.MatchesData32(topic) {
			return false
		}
	}

	return true
}

func (b *Bloom) MatchesAddress(addr Address) bool {
	return b.MatchesBytes(addr.Bytes())
}

func (b *Bloom) MatchesData32(data Data32) bool {
	return b.MatchesBytes(data.Bytes())
}

func (b *Bloom) MatchesBytes(_bytes []byte) bool {
	for _, bits := range b.bloomBits(_bytes) {
		if !b.has(bits) {
			return false
		}
	}
	return true
}

func (b *Bloom) bloomBits(_bytes []byte) []int {
	// Inspired by https://github.com/hyperdivision/eth-bloomfilter/blob/master/index.js#L3
	hash := sha3.NewLegacyKeccak256()
	hash.Write(_bytes)
	d := hash.Sum(nil)
	return []int{
		toBit(d[0], d[1]),
		toBit(d[2], d[3]),
		toBit(d[4], d[5]),
	}
}

func (b *Bloom) set(pos int) {
	b.value[255-(pos>>3)] |= byte(1 << byte(pos&7))
}

func (b *Bloom) has(pos int) bool {
	return (b.value[255-(pos>>3)] & byte(1<<byte(pos&7))) != 0
}

func toBit(high, low byte) int {
	h := int(high)
	l := int(low)
	return ((h << 8) + l) & 2047
}
