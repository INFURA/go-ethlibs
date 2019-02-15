package eth

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

type Quantity struct {
	s string
	i big.Int
}

func MustQuantity(value string) *Quantity {
	q, err := NewQuantity(value)
	if err != nil {
		panic(err)
	}
	return q
}

func NewQuantity(value string) (*Quantity, error) {
	q := Quantity{}
	// Save the string
	q.s = value

	// If the hex string is odd assume it's because a leading zero was removed
	if len(value)%2 != 0 {
		value = "0x0" + value[2:]
	}

	b, err := hex.DecodeString(value[2:])
	if err != nil {
		return nil, err
	}

	q.i.SetBytes(b)
	return &q, nil
}

func QuantityFromInt64(value int64) Quantity {
	return Quantity{
		s: "",
		i: *big.NewInt(value),
	}
}

func QuantityFromUInt64(value uint64) Quantity {
	return Quantity{
		s: "",
		i: *big.NewInt(0).SetUint64(value),
	}
}

func (q *Quantity) UnmarshalJSON(data []byte) error {
	str, err := unmarshalHex(data, 0, "quantity")
	if err != nil {
		return err
	}

	_q, err := NewQuantity(str)
	if err != nil {
		return err
	}

	*q = *_q
	return nil
}

func (q *Quantity) MarshalJSON() ([]byte, error) {
	if q.s != "" {
		return json.Marshal(&q.s)
	}

	b := q.i.Bytes()
	if len(b) == 0 {
		// If we are a 0 value Quantity, make sure we return 0x0 and not 0x.
		s := "0x0"
		return json.Marshal(&s)
	}

	h := hex.EncodeToString(b)

	// remove any leading 0s
	h = strings.TrimLeft(h, "0")
	s := fmt.Sprintf("0x%s", h)
	return json.Marshal(&s)
}

func (q *Quantity) String() string {
	return q.s
}

func (q *Quantity) UInt64() uint64 {
	return q.i.Uint64()
}

func (q *Quantity) Int64() int64 {
	return q.i.Int64()
}

func (q *Quantity) Big() *big.Int {
	return &q.i
}
