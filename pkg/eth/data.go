package eth

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

type Data string
type Data8 Data
type Data20 Data
type Data32 Data
type Data256 Data

// Aliases
type Hash = Data32
type Topic = Data32

func (d Data) String() string {
	return string(d)
}

func (d Data8) String() string {
	return string(d)
}

func (d Data20) String() string {
	return string(d)
}

func (d Data32) String() string {
	return string(d)
}

func (d Data256) String() string {
	return string(d)
}

func (d *Data) UnmarshalJSON(data []byte) error {
	str, err := unmarshalHex(data, 0, "data")
	if err != nil {
		return err
	}
	*d = Data(str)
	return nil
}

func (d *Data8) UnmarshalJSON(data []byte) error {
	str, err := unmarshalHex(data, 8, "data")
	if err != nil {
		return err
	}
	*d = Data8(str)
	return nil
}

func (d *Data20) UnmarshalJSON(data []byte) error {
	str, err := unmarshalHex(data, 20, "data")
	if err != nil {
		return err
	}
	*d = Data20(str)
	return nil
}

func (d *Data32) UnmarshalJSON(data []byte) error {
	str, err := unmarshalHex(data, 32, "data")
	if err != nil {
		return err
	}
	*d = Data32(str)
	return nil
}

func (d *Data256) UnmarshalJSON(data []byte) error {
	str, err := unmarshalHex(data, 256, "data")
	if err != nil {
		return err
	}
	*d = Data256(str)
	return nil
}

func (d *Data) MarshalJSON() ([]byte, error) {
	s := string(*d)
	return json.Marshal(&s)
}

func unmarshalHex(data []byte, size int, typ string) (string, error) {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(str, "0x") {
		return "", errors.Errorf("%s types must start with 0x", typ)
	}

	if size != 0 {
		dataSize := (len(str) - 2) / 2

		if size != dataSize {
			return "", errors.Errorf("%s type size mismatch, expected %d got %d", typ, size, dataSize)
		}
	}

	return str, nil
}
