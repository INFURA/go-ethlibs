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

	// EIP-1559 BaseFeePerGas
	BaseFeePerGas *Quantity `json:"baseFeePerGas,omitempty"`

	// EIP-4895 Withdrawals
	WithdrawalsRoot *Data32      `json:"withdrawalsRoot,omitempty"`
	Withdrawals     []Withdrawal `json:"withdrawals,omitempty"`

	// EIP-4788 Beacon Block Root
	ParentBeaconBlockRoot *Hash `json:"parentBeaconBlockRoot,omitempty"`

	// EIP-4844 Blob related block fields
	ExcessBlobGas *Quantity `json:"excessBlobGas,omitempty"`
	BlobGasUsed   *Quantity `json:"blobGasUsed,omitempty"`

	// Ethhash POW Fields
	Nonce   *Data8 `json:"nonce"`
	MixHash *Data  `json:"mixHash"`

	// POA Fields (Aura)
	Step      *string `json:"step,omitempty"`
	Signature *string `json:"signature,omitempty"`

	// Parity Specific Fields
	SealFields *[]Data `json:"sealFields,omitempty"`

	// Track the flavor so we can re-encode correctly
	flavor string
}

func (b *Block) DepopulateTransactions() {
	for i := range b.Transactions {
		b.Transactions[i].Populated = false
	}
}

func (b *Block) UnmarshalJSON(data []byte) error {
	type block Block
	aliased := block(*b)

	err := json.Unmarshal(data, &aliased)
	if err != nil {
		return err
	}

	*b = Block(aliased)
	if b.SealFields == nil {
		// It's a geth response, which is always the same regardless of consensus algorithm
		b.flavor = "geth"
	} else {
		// It's a parity response, which differs by the consensus algorithm
		b.flavor = "parity-unknown"

		if b.MixHash != nil {
			// it's ethhash
			b.flavor = "parity-ethhash"
		} else if b.Step != nil && b.Signature != nil {
			// it's Aura-based POA
			b.flavor = "parity-aura"
		} else {
			// it's Clique-based POA
			b.flavor = "parity-clique"
		}
	}
	return nil
}

func (b Block) MarshalJSON() ([]byte, error) {
	switch b.flavor {
	case "geth":
		// Note that post EIP-4895 we have to behave slightly differently if WithdrawalsRoot is set.
		// if it's set, then `Withdrawals` should be non-nil
		// if it's not set, then Withdrawals should be omitted
		if b.WithdrawalsRoot != nil {
			type withWithdrawals struct {
				Number           *Quantity  `json:"number"`
				Hash             *Hash      `json:"hash"`
				ParentHash       Hash       `json:"parentHash"`
				SHA3Uncles       Data32     `json:"sha3Uncles"`
				LogsBloom        Data256    `json:"logsBloom"`
				TransactionsRoot Data32     `json:"transactionsRoot"`
				StateRoot        Data32     `json:"stateRoot"`
				ReceiptsRoot     Data32     `json:"receiptsRoot"`
				Miner            Address    `json:"miner"`
				Difficulty       Quantity   `json:"difficulty"`
				TotalDifficulty  Quantity   `json:"totalDifficulty"`
				ExtraData        Data       `json:"extraData"`
				Size             Quantity   `json:"size"`
				GasLimit         Quantity   `json:"gasLimit"`
				GasUsed          Quantity   `json:"gasUsed"`
				Timestamp        Quantity   `json:"timestamp"`
				Transactions     []TxOrHash `json:"transactions"`
				Uncles           []Hash     `json:"uncles"`

				// EIP-1559 BaseFeePerGas
				BaseFeePerGas *Quantity `json:"baseFeePerGas,omitempty"`

				// EIP-4895 Withdrawals
				WithdrawalsRoot *Data32      `json:"withdrawalsRoot"`
				Withdrawals     []Withdrawal `json:"withdrawals"`

				// EIP-4788 Beacon Block Root
				ParentBeaconBlockRoot *Hash `json:"parentBeaconBlockRoot,omitempty"`

				// EIP-4844 Blob related block fields
				ExcessBlobGas *Quantity `json:"excessBlobGas,omitempty"`
				BlobGasUsed   *Quantity `json:"blobGasUsed,omitempty"`

				Nonce   *Data8 `json:"nonce"`
				MixHash *Data  `json:"mixHash"`
			}

			w := withWithdrawals{
				Number:                b.Number,
				Hash:                  b.Hash,
				ParentHash:            b.ParentHash,
				SHA3Uncles:            b.SHA3Uncles,
				LogsBloom:             b.LogsBloom,
				TransactionsRoot:      b.TransactionsRoot,
				StateRoot:             b.StateRoot,
				ReceiptsRoot:          b.ReceiptsRoot,
				Miner:                 b.Miner,
				Difficulty:            b.Difficulty,
				TotalDifficulty:       b.TotalDifficulty,
				ExtraData:             b.ExtraData,
				Size:                  b.Size,
				GasLimit:              b.GasLimit,
				GasUsed:               b.GasUsed,
				Timestamp:             b.Timestamp,
				Transactions:          b.Transactions,
				Uncles:                b.Uncles,
				BaseFeePerGas:         b.BaseFeePerGas,
				Nonce:                 b.Nonce,
				MixHash:               b.MixHash,
				WithdrawalsRoot:       b.WithdrawalsRoot,
				Withdrawals:           b.Withdrawals,
				ParentBeaconBlockRoot: b.ParentBeaconBlockRoot,
				ExcessBlobGas:         b.ExcessBlobGas,
				BlobGasUsed:           b.BlobGasUsed,
			}

			return json.Marshal(&w)
		}

		// otherwise, fallback to pre-EIP-4895 behavior
		type geth struct {
			Number           *Quantity  `json:"number"`
			Hash             *Hash      `json:"hash"`
			ParentHash       Hash       `json:"parentHash"`
			SHA3Uncles       Data32     `json:"sha3Uncles"`
			LogsBloom        Data256    `json:"logsBloom"`
			TransactionsRoot Data32     `json:"transactionsRoot"`
			StateRoot        Data32     `json:"stateRoot"`
			ReceiptsRoot     Data32     `json:"receiptsRoot"`
			Miner            Address    `json:"miner"`
			Difficulty       Quantity   `json:"difficulty"`
			TotalDifficulty  Quantity   `json:"totalDifficulty"`
			ExtraData        Data       `json:"extraData"`
			Size             Quantity   `json:"size"`
			GasLimit         Quantity   `json:"gasLimit"`
			GasUsed          Quantity   `json:"gasUsed"`
			Timestamp        Quantity   `json:"timestamp"`
			Transactions     []TxOrHash `json:"transactions"`
			Uncles           []Hash     `json:"uncles"`

			// EIP-1559 BaseFeePerGas
			BaseFeePerGas *Quantity `json:"baseFeePerGas,omitempty"`

			Nonce   *Data8 `json:"nonce"`
			MixHash *Data  `json:"mixHash"`
		}

		g := geth{
			Number:           b.Number,
			Hash:             b.Hash,
			ParentHash:       b.ParentHash,
			SHA3Uncles:       b.SHA3Uncles,
			LogsBloom:        b.LogsBloom,
			TransactionsRoot: b.TransactionsRoot,
			StateRoot:        b.StateRoot,
			ReceiptsRoot:     b.ReceiptsRoot,
			Miner:            b.Miner,
			Difficulty:       b.Difficulty,
			TotalDifficulty:  b.TotalDifficulty,
			ExtraData:        b.ExtraData,
			Size:             b.Size,
			GasLimit:         b.GasLimit,
			GasUsed:          b.GasUsed,
			Timestamp:        b.Timestamp,
			Transactions:     b.Transactions,
			Uncles:           b.Uncles,
			BaseFeePerGas:    b.BaseFeePerGas,
			Nonce:            b.Nonce,
			MixHash:          b.MixHash,
		}

		return json.Marshal(&g)
	case "parity-aura":
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
			SealFields:       b.SealFields,
			Step:             b.Step,
			Signature:        b.Signature,
		}

		return json.Marshal(&a)
	case "parity-clique":
		type clique struct {
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

			SealFields *[]Data `json:"sealFields,omitempty"`
		}

		c := clique{
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
			SealFields:       b.SealFields,
		}

		return json.Marshal(&c)
	case "parity-ethhash":
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

			Nonce      *Data8  `json:"nonce"`
			MixHash    *Data   `json:"mixHash"`
			SealFields *[]Data `json:"sealFields,omitempty"`
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
			Nonce:            b.Nonce,
			MixHash:          b.MixHash,
			SealFields:       b.SealFields,
		}

		return json.Marshal(&e)
	}

	type unknown Block
	u := unknown(b)
	return json.Marshal(&u)
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

func (t TxOrHash) MarshalJSON() ([]byte, error) {
	if t.Populated {
		return json.Marshal(&t.Transaction)
	}

	return json.Marshal(&t.Hash)
}
