package eth

import (
	"encoding/json"
	"strings"

	"github.com/INFURA/go-ethlibs/rlp"
	"github.com/pkg/errors"
)

type Input Data

func NewInput(value string) (*Input, error) {
	if !strings.HasPrefix(value, "0x") {
		return nil, errors.Errorf("invalid input: %s", value)
	}

	a := Input(value)
	return &a, nil
}

func MustInput(value string) *Input {
	a, err := NewInput(value)
	if err != nil {
		panic(err)
	}

	return a
}

func (i Input) String() string {
	return string(i)
}

func (i Input) Bytes() []byte {
	return Data(i).Bytes()
}

// RLP returns the Input as an RLP-encoded string, note Input can never be null
func (i Input) RLP() rlp.Value {
	return rlp.Value{
		String: strings.ToLower(i.String()),
	}
}

func (i Input) FunctionSelector() *Data4 {
	if len(i) >= 10 {
		b := Data4(i[:10])
		return &b
	}

	return nil
}

func (i *Input) UnmarshalJSON(data []byte) error {
	str, err := unmarshalHex(data, -1, "data")
	if err != nil {
		return err
	}
	*i = Input(str)
	return nil
}

func (i Input) MarshalJSON() ([]byte, error) {
	s := string(i)
	return json.Marshal(&s)
}
