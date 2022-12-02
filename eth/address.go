package eth

import (
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"

	"github.com/ConsenSys/go-ethlibs/rlp"
)

type Address Data20

func NewAddress(value string) (*Address, error) {
	// 20 bytes plus 0x
	if len(value) != 42 || !strings.HasPrefix(value, "0x") {
		return nil, errors.Errorf("invalid address: %s", value)
	}

	a := Address(ToChecksumAddress(value))
	return &a, nil
}

func MustAddress(value string) *Address {
	a, err := NewAddress(value)
	if err != nil {
		panic(err)
	}

	return a
}

func (a Address) String() string {
	return string(a)
}

func (a Address) Bytes() []byte {
	return Data20(a).Bytes()
}

func (a *Address) UnmarshalJSON(data []byte) error {
	// We'll keep the checksummed string in memory so we can use it for internal representations
	str, err := unmarshalHex(data, 20, "data")
	str = ToChecksumAddress(str)
	if err != nil {
		return err
	}
	*a = Address(str)
	return nil
}

func (a Address) MarshalJSON() ([]byte, error) {
	// Seems like geth and parity both return the lower-cased string rather than the checksummed one
	s := strings.ToLower(string(a))
	return json.Marshal(&s)
}

// RLP returns the Address as an RLP-encoded string, or an empty RLP string for the nil Address.
func (a *Address) RLP() rlp.Value {
	if a == nil {
		return rlp.Value{
			String: "0x",
		}
	}
	return rlp.Value{
		String: strings.ToLower(a.String()),
	}
}

/*
ToChecksumAddress converts a string to the proper EIP55 casing.

Transliteration of this code from the EIP55 wiki page:

	function toChecksumAddress (address) {
	  address = address.toLowerCase().replace('0x', '')
	  var hash = createKeccakHash('keccak256').update(address).digest('hex')
	  var ret = '0x'

	  for (var i = 0; i < address.length; i++) {
		if (parseInt(hash[i], 16) >= 8) {
		  ret += address[i].toUpperCase()
		} else {
		  ret += address[i]
		}
	  }

	  return ret
	}
*/
func ToChecksumAddress(address string) string {
	address = strings.Replace(strings.ToLower(address), "0x", "", 1)
	hash := sha3.NewLegacyKeccak256()
	_, _ = hash.Write([]byte(address))
	sum := hash.Sum(nil)
	digest := hex.EncodeToString(sum)

	b := strings.Builder{}
	b.WriteString("0x")

	for i := 0; i < len(address); i++ {
		a := address[i]
		if a > '9' {
			d, _ := strconv.ParseInt(digest[i:i+1], 16, 8)

			if d >= 8 {
				// Upper case it
				a -= 'a' - 'A'
				b.WriteByte(a)
			} else {
				// Keep it lower
				b.WriteByte(a)
			}
		} else {
			// Keep it lower
			b.WriteByte(a)
		}
	}

	return b.String()
}
