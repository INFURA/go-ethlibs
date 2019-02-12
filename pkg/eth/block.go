package eth

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

type Block struct {
	Number           BlockNumber `json:"number"`
	Hash             *Hash       `json:"hash"`
	ParentHash       *Hash       `json:"parentHash"`
	Nonce            *Data8      `json:"nonce"`
	SHA3Uncles       Data32      `json:"sha3Uncles"`
	LogsBloom        Data256     `json:"logsBloom"`
	TransactionsRoot Data32      `json:"transactionsRoot"`
	StateRoot        Data32      `json:"stateRoot"`
	ReceiptsRoot     Data32      `json:"receiptsRoot"`
	Miner            Address     `json:"miner"`
	Author           Address     `json:"author"` // Parity-specific alias of miner
	Difficulty       Quantity    `json:"difficulty"`
	TotalDifficulty  Quantity    `json:"totalDifficulty"`
	ExtraData        Data        `json:"extraData"`
	Size             Quantity    `json:"size"`
	GasLimit         Quantity    `json:"gasLimit"`
	GasUsed          Quantity    `json:"gasUsed"`
	Timestamp        Quantity    `json:"timestamp"`
	Transactions     []TxOrHash  `json:"transactions"`
	Uncles           []Hash      `json:"uncles"`

	// POA Fields (Aura)
	SealFields []Data `json:"sealFields,omitempty"`
	Step       string `json:"step,omitempty"`
	Signature  string `json:"signature,omitempty"`

	// POA Fields (Clique)
	MixHash *Data `json:"mixHash,omitempty"`

	// Track the source so we can re-encode correctly
	source string `json:"-"`
}

type BlockNumber struct {
	i int64
	s string
}

func (bn *BlockNumber) UnmarshalJSON(data []byte) error {
	str, err := unmarshalHex(data, 0, "quantity")
	if err != nil {
		return err
	}

	// Save the string
	bn.s = str

	// If the hex string is odd assume it's because a leading zero was removed
	if len(str)%2 != 0 {
		str = "0x0" + str[2:]
	}

	b, err := hex.DecodeString(str[2:])
	if err != nil {
		return err
	}

	i := big.Int{}
	i.SetBytes(b)

	bn.i = i.Int64()
	return nil
}

func (bn *BlockNumber) MarshalJSON() ([]byte, error) {
	if bn.s != "" {
		return json.Marshal(&bn.s)
	}

	i := big.NewInt(bn.i)
	b := i.Bytes()
	h := hex.EncodeToString(b)

	// remove any leading 0s
	h = strings.TrimLeft(h, "0")
	s := fmt.Sprintf("0x%s", h)
	return json.Marshal(&s)
}

func (bn *BlockNumber) String() string {
	return bn.s
}

func (bn *BlockNumber) Int64() int64 {
	return bn.i
}

type TxOrHash struct {
	Transaction
	Populated bool `json:"-"`
}

func (t *TxOrHash) UnmarshalJSON(data []byte) error {
	// if input is just a string, then it's a hash, if it's an object it's a "full" transaction
	err := json.Unmarshal(data, &t.Hash)
	if err == nil {
		t.Populated = false
		return nil
	}

	t.Populated = true
	return json.Unmarshal(data, &t.Transaction)

}

func (t *TxOrHash) MarshalJSON() ([]byte, error) {
	if t.Populated {
		return json.Marshal(&t.Transaction)
	}

	return json.Marshal(&t.Hash)
}
