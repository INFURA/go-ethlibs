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

func (q *Quantity) UnmarshalJSON(data []byte) error {
	str, err := unmarshalHex(data, 0, "quantity")
	if err != nil {
		return err
	}

	// Save the string
	q.s = str

	// If the hex string is odd assume it's because a leading zero was removed
	if len(str)%2 != 0 {
		str = "0x0" + str[2:]
	}

	b, err := hex.DecodeString(str[2:])
	if err != nil {
		return err
	}

	q.i.SetBytes(b)
	return nil
}

func (q *Quantity) MarshalJSON() ([]byte, error) {
	if q.s != "" {
		return json.Marshal(&q.s)
	}

	b := q.i.Bytes()
	h := hex.EncodeToString(b)

	// remove any leading 0s
	h = strings.TrimLeft(h, "0")
	s := fmt.Sprintf("0x%s", h)
	return json.Marshal(&s)
}

func (q *Quantity) UInt64() uint64 {
	return q.i.Uint64()
}

func (q *Quantity) Int64() int64 {
	return q.i.Int64()
}

func (q *Quantity) Big() big.Int {
	return q.i
}
