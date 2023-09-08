package eth

import (
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"

	"github.com/INFURA/go-ethlibs/rlp"
)

type Data string
type Data8 Data
type Data20 Data
type Data32 Data
type Data256 Data

// Aliases
type Hash = Data32
type Topic = Data32

func NewData(value string) (*Data, error) {
	parsed, err := validateHex(value, -1, "data")
	if err != nil {
		return nil, err
	}

	d := Data(parsed)
	return &d, nil
}

func NewData8(value string) (*Data8, error) {
	parsed, err := validateHex(value, 8, "data")
	if err != nil {
		return nil, err
	}

	d := Data8(parsed)
	return &d, nil
}

func NewData20(value string) (*Data20, error) {
	parsed, err := validateHex(value, 20, "data")
	if err != nil {
		return nil, err
	}

	d := Data20(parsed)
	return &d, nil
}

func NewData32(value string) (*Data32, error) {
	parsed, err := validateHex(value, 32, "data")
	if err != nil {
		return nil, err
	}

	d := Data32(parsed)
	return &d, nil
}

func NewHash(value string) (*Hash, error) {
	return NewData32(value)
}

func NewTopic(value string) (*Hash, error) {
	return NewData32(value)
}

func NewData256(value string) (*Data256, error) {
	parsed, err := validateHex(value, 256, "data")
	if err != nil {
		return nil, err
	}

	d := Data256(parsed)
	return &d, nil
}

func MustData(value string) *Data {
	d, err := NewData(value)
	if err != nil {
		panic(err)
	}

	return d
}

func MustData8(value string) *Data8 {
	d, err := NewData8(value)
	if err != nil {
		panic(err)
	}

	return d
}

func MustData20(value string) *Data20 {
	d, err := NewData20(value)
	if err != nil {
		panic(err)
	}

	return d
}

func MustData32(value string) *Data32 {
	d, err := NewData32(value)
	if err != nil {
		panic(err)
	}

	return d
}

func MustHash(value string) *Hash {
	return MustData32(value)
}

func MustTopic(value string) *Hash {
	return MustData32(value)
}

func MustData256(value string) *Data256 {
	d, err := NewData256(value)
	if err != nil {
		panic(err)
	}

	return d
}

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

func (d Data) Bytes() []byte {
	b, err := hex.DecodeString(d.String()[2:])
	if err != nil {
		panic(err)
	}
	return b
}

func (d Data8) Bytes() []byte {
	b, err := hex.DecodeString(d.String()[2:])
	if err != nil {
		panic(err)
	}
	return b
}

func (d Data20) Bytes() []byte {
	b, err := hex.DecodeString(d.String()[2:])
	if err != nil {
		panic(err)
	}
	return b
}

func (d Data32) Bytes() []byte {
	b, err := hex.DecodeString(d.String()[2:])
	if err != nil {
		panic(err)
	}
	return b
}

func (d Data256) Bytes() []byte {
	b, err := hex.DecodeString(d.String()[2:])
	if err != nil {
		panic(err)
	}
	return b
}

// Hash returns the keccak256 hash of the Data.
func (d Data) Hash() Hash {
	return hash(d)
}

// Hash returns the keccak256 hash of the Data8.
func (d Data8) Hash() Hash {
	return hash(d)
}

// Hash returns the keccak256 hash of the Data20.
func (d Data20) Hash() Hash {
	return hash(d)
}

// Hash returns the keccak256 hash of the Data32.
func (d Data32) Hash() Hash {
	return hash(d)
}

// Hash returns the keccak256 hash of the Data256.
func (d Data256) Hash() Hash {
	return hash(d)
}

func (d *Data) UnmarshalJSON(data []byte) error {
	str, err := unmarshalHex(data, -1, "data")
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

func (d Data) MarshalJSON() ([]byte, error) {
	s := string(d)
	return json.Marshal(&s)
}

func unmarshalHex(data []byte, size int, typ string) (string, error) {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return "", err
	}

	return validateHex(str, size, typ)
}

func validateHex(value string, size int, typ string) (string, error) {
	if !strings.HasPrefix(value, "0x") {
		return "", errors.Errorf("%s types must start with 0x", typ)
	}

	if size != -1 {
		dataSize := (len(value) - 2) / 2

		if size != dataSize {
			return "", errors.Errorf("%s type size mismatch, expected %d got %d", typ, size, dataSize)
		}
	}

	// validate that the input characters after 0x are only 0-9, a-f, A-F
	for i, c := range value[2:] {
		switch {
		case '0' <= c && c <= '9':
			continue
		case 'a' <= c && c <= 'f':
			continue
		case 'A' <= c && c <= 'F':
			continue
		}

		return "", errors.Errorf("invalid hex string, invalid character '%c' at index %d", c, i+2)
	}

	return value, nil
}

// RLP returns the Data as an RLP-encoded string.
func (d *Data) RLP() rlp.Value {
	return rlp.Value{
		String: d.String(),
	}
}

// RLP returns the Data8 as an RLP-encoded string.
func (d *Data8) RLP() rlp.Value {
	return rlp.Value{
		String: d.String(),
	}
}

// RLP returns the Data32 as an RLP-encoded string.
func (d *Data32) RLP() rlp.Value {
	return rlp.Value{
		String: d.String(),
	}
}

// RLP returns the Data256 as an RLP-encoded string.
func (d *Data256) RLP() rlp.Value {
	return rlp.Value{
		String: d.String(),
	}
}

type Hashes []Data32

func (slice Hashes) RLP() rlp.Value {
	v := rlp.Value{
		List: make([]rlp.Value, len(slice)),
	}
	for i := range slice {
		v.List[i].String = slice[i].String()
	}
	return v
}

type hasBytes interface {
	Bytes() []byte
}

func hash(from hasBytes) Hash {
	b := from.Bytes()
	// And feed the bytes into our hash
	hash := sha3.NewLegacyKeccak256()
	hash.Write(b)
	sum := hash.Sum(nil)

	// and finally return the hash as a 0x prefixed string
	digest := hex.EncodeToString(sum)
	return Hash("0x" + digest)
}
