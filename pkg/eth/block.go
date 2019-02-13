package eth

import (
	"encoding/json"
)

type Block struct {
	Number           *Quantity  `json:"number"`
	Hash             *Hash      `json:"hash"`
	ParentHash       Hash       `json:"parentHash"`
	SHA3Uncles       Data32     `json:"sha3Uncles"`
	LogsBloom        Data256    `json:"logsBloom"`
	TransactionsRoot Data32     `json:"transactionsRoot"`
	StateRoot        Data32     `json:"stateRoot"`
	ReceiptsRoot     Data32     `json:"receiptsRoot"`
	Miner            Address    `json:"miner"`
	Author           Address    `json:"author,omitempty"` // Parity-specific alias of miner
	Difficulty       Quantity   `json:"difficulty"`
	TotalDifficulty  Quantity   `json:"totalDifficulty"`
	ExtraData        Data       `json:"extraData"`
	Size             Quantity   `json:"size"`
	GasLimit         Quantity   `json:"gasLimit"`
	GasUsed          Quantity   `json:"gasUsed"`
	Timestamp        Quantity   `json:"timestamp"`
	Transactions     []TxOrHash `json:"transactions"`
	Uncles           []Hash     `json:"uncles"`

	// Ethhash POW Fields
	Nonce   *Data8 `json:"nonce"`
	MixHash *Data  `json:"mixHash"`

	// POA Fields (Aura)
	SealFields *[]Data `json:"sealFields,omitempty"`
	Step       *string `json:"step,omitempty"`
	Signature  *string `json:"signature,omitempty"`

	// Track the source so we can re-encode correctly
	source string `json:"-"`
}

func (b *Block) UnmarshalJSON(data []byte) error {
	type block Block
	aliased := block(*b)

	err := json.Unmarshal(data, &aliased)
	if err != nil {
		return err
	}

	*b = Block(aliased)
	if b.MixHash != nil {
		b.source = "ethhash"
	} else if b.Step != nil {
		b.source = "aura"
	}

	return nil
}

func (b *Block) MarshalJSON() ([]byte, error) {
	if b.source == "ethhash" {
		type ethhash struct {
			Number           *Quantity  `json:"number"`
			Hash             *Hash      `json:"hash"`
			ParentHash       Hash       `json:"parentHash"`
			SHA3Uncles       Data32     `json:"sha3Uncles"`
			LogsBloom        Data256    `json:"logsBloom"`
			TransactionsRoot Data32     `json:"transactionsRoot"`
			StateRoot        Data32     `json:"stateRoot"`
			ReceiptsRoot     Data32     `json:"receiptsRoot"`
			Miner            Address    `json:"miner"`
			Author           Address    `json:"author,omitempty"` // Parity-specific alias of miner
			Difficulty       Quantity   `json:"difficulty"`
			TotalDifficulty  Quantity   `json:"totalDifficulty"`
			ExtraData        Data       `json:"extraData"`
			Size             Quantity   `json:"size"`
			GasLimit         Quantity   `json:"gasLimit"`
			GasUsed          Quantity   `json:"gasUsed"`
			Timestamp        Quantity   `json:"timestamp"`
			Transactions     []TxOrHash `json:"transactions"`
			Uncles           []Hash     `json:"uncles"`

			// Ethhash POW Fields
			Nonce   *Data8 `json:"nonce"`
			MixHash *Data  `json:"mixHash"`
		}

		e := ethhash{
			Number:           b.Number,
			Hash:             b.Hash,
			ParentHash:       b.ParentHash,
			SHA3Uncles:       b.SHA3Uncles,
			LogsBloom:        b.LogsBloom,
			TransactionsRoot: b.TransactionsRoot,
			StateRoot:        b.StateRoot,
			ReceiptsRoot:     b.ReceiptsRoot,
			Miner:            b.Miner,
			Author:           b.Author,
			Difficulty:       b.Difficulty,
			TotalDifficulty:  b.TotalDifficulty,
			ExtraData:        b.ExtraData,
			Size:             b.Size,
			GasLimit:         b.GasLimit,
			GasUsed:          b.GasUsed,
			Timestamp:        b.Timestamp,
			Transactions:     b.Transactions,
			Uncles:           b.Uncles,

			// Ethhash POW Fields
			Nonce:   b.Nonce,
			MixHash: b.MixHash,
		}

		return json.Marshal(&e)
	} else if b.source == "aura" {
		type aura struct {
			Number           *Quantity  `json:"number"`
			Hash             *Hash      `json:"hash"`
			ParentHash       Hash       `json:"parentHash"`
			SHA3Uncles       Data32     `json:"sha3Uncles"`
			LogsBloom        Data256    `json:"logsBloom"`
			TransactionsRoot Data32     `json:"transactionsRoot"`
			StateRoot        Data32     `json:"stateRoot"`
			ReceiptsRoot     Data32     `json:"receiptsRoot"`
			Miner            Address    `json:"miner"`
			Author           Address    `json:"author,omitempty"` // Parity-specific alias of miner
			Difficulty       Quantity   `json:"difficulty"`
			TotalDifficulty  Quantity   `json:"totalDifficulty"`
			ExtraData        Data       `json:"extraData"`
			Size             Quantity   `json:"size"`
			GasLimit         Quantity   `json:"gasLimit"`
			GasUsed          Quantity   `json:"gasUsed"`
			Timestamp        Quantity   `json:"timestamp"`
			Transactions     []TxOrHash `json:"transactions"`
			Uncles           []Hash     `json:"uncles"`

			// POA Fields (Aura)
			SealFields *[]Data `json:"sealFields,omitempty"`
			Step       *string `json:"step,omitempty"`
			Signature  *string `json:"signature,omitempty"`
		}

		a := aura{
			Number:           b.Number,
			Hash:             b.Hash,
			ParentHash:       b.ParentHash,
			SHA3Uncles:       b.SHA3Uncles,
			LogsBloom:        b.LogsBloom,
			TransactionsRoot: b.TransactionsRoot,
			StateRoot:        b.StateRoot,
			ReceiptsRoot:     b.ReceiptsRoot,
			Miner:            b.Miner,
			Author:           b.Author,
			Difficulty:       b.Difficulty,
			TotalDifficulty:  b.TotalDifficulty,
			ExtraData:        b.ExtraData,
			Size:             b.Size,
			GasLimit:         b.GasLimit,
			GasUsed:          b.GasUsed,
			Timestamp:        b.Timestamp,
			Transactions:     b.Transactions,
			Uncles:           b.Uncles,

			// POA Fields (Aura)
			SealFields: b.SealFields,
			Step:       b.Step,
			Signature:  b.Signature,
		}

		return json.Marshal(&a)
	}

	type unknown Block
	u := unknown(*b)
	return json.Marshal(&u)
}

/*
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
*/

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
