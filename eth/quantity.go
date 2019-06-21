package eth

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/pkg/errors"
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

	if !strings.HasPrefix(value, "0x") {
		return nil, errors.New("quantity values must start with 0x")
	}

	if value == "0x" {
		return nil, errors.New("quantity values must include at least one digit")
	}

	if strings.HasPrefix(value, "0x0") && value != "0x0" {
		return nil, errors.New("quantity values should not have leading zeroes")
	}

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
	i := big.NewInt(value)
	return Quantity{
		s: bigToQuantityString(i),
		i: *i,
	}
}

func QuantityFromUInt64(value uint64) Quantity {
	i := big.NewInt(0).SetUint64(value)
	return Quantity{
		s: bigToQuantityString(i),
		i: *i,
	}
}

func QuantityFromBigInt(value *big.Int) Quantity {
	return Quantity{
		s: bigToQuantityString(value),
		i: *value,
	}
}

func (q *Quantity) UnmarshalJSON(data []byte) error {
	str, err := unmarshalHex(data, -1, "quantity")
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
	if q.s == "" {
		q.s = bigToQuantityString(&q.i)
	}

	return json.Marshal(&q.s)
}

func (q Quantity) String() string {
	return q.s
}

func (q Quantity) UInt64() uint64 {
	return q.i.Uint64()
}

func (q Quantity) Int64() int64 {
	return q.i.Int64()
}

func (q Quantity) Big() *big.Int {
	return &q.i
}

func bigToQuantityString(i *big.Int) string {
	b := i.Bytes()
	if len(b) == 0 {
		// If we are a 0 value Quantity, make sure we return 0x0 and not 0x.
		return "0x0"
	}

	h := hex.EncodeToString(b)

	// remove any leading 0s
	h = strings.TrimLeft(h, "0")
	return fmt.Sprintf("0x%s", h)
}
