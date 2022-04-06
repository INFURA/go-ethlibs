package eth

import "encoding/json"

type Uncle struct {
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
	TotalDifficulty  *Quantity  `json:"totalDifficulty"` // null when using Geth getUncle APIs
	ExtraData        Data       `json:"extraData"`
	Size             *Quantity  `json:"size"` // null when using parity getUncle APIs
	GasLimit         Quantity   `json:"gasLimit"`
	GasUsed          Quantity   `json:"gasUsed"`
	Timestamp        Quantity   `json:"timestamp"`
	Transactions     []TxOrHash `json:"transactions"` // missing from geth, always empty in parity
	Uncles           []Hash     `json:"uncles"`

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

func (u *Uncle) UnmarshalJSON(data []byte) error {
	type uncle Uncle
	aliased := uncle(*u)

	err := json.Unmarshal(data, &aliased)
	if err != nil {
		return err
	}

	*u = Uncle(aliased)
	if u.SealFields == nil {
		// It's a geth response, which is always the same regardless of consensus algorithm
		u.flavor = "geth"
	} else {
		// It's a parity response, which differs by the consensus algorithm
		u.flavor = "parity-unknown"

		if u.MixHash != nil {
			// it's ethhash
			u.flavor = "parity-ethhash"
		} else if u.Step != nil && u.Signature != nil {
			// it's Aura-based POA
			u.flavor = "parity-aura"
		} else {
			// it's Clique-based POA
			u.flavor = "parity-clique"
		}
	}
	return nil
}

func (u Uncle) MarshalJSON() ([]byte, error) {
	switch u.flavor {
	case "geth":
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
			TotalDifficulty  *Quantity  `json:"totalDifficulty"`
			ExtraData        Data       `json:"extraData"`
			Size             *Quantity  `json:"size"`
			GasLimit         Quantity   `json:"gasLimit"`
			GasUsed          Quantity   `json:"gasUsed"`
			Timestamp        Quantity   `json:"timestamp"`
			Transactions     []TxOrHash `json:"transactions,omitempty"`
			Uncles           []Hash     `json:"uncles"`

			Nonce   *Data8 `json:"nonce"`
			MixHash *Data  `json:"mixHash"`
		}

		g := geth{
			Number:           u.Number,
			Hash:             u.Hash,
			ParentHash:       u.ParentHash,
			SHA3Uncles:       u.SHA3Uncles,
			LogsBloom:        u.LogsBloom,
			TransactionsRoot: u.TransactionsRoot,
			StateRoot:        u.StateRoot,
			ReceiptsRoot:     u.ReceiptsRoot,
			Miner:            u.Miner,
			Difficulty:       u.Difficulty,
			TotalDifficulty:  u.TotalDifficulty,
			ExtraData:        u.ExtraData,
			Size:             u.Size,
			GasLimit:         u.GasLimit,
			GasUsed:          u.GasUsed,
			Timestamp:        u.Timestamp,
			Transactions:     u.Transactions,
			Uncles:           u.Uncles,
			Nonce:            u.Nonce,
			MixHash:          u.MixHash,
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
			TotalDifficulty  *Quantity  `json:"totalDifficulty"`
			ExtraData        Data       `json:"extraData"`
			Size             *Quantity  `json:"size"`
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
			Number:           u.Number,
			Hash:             u.Hash,
			ParentHash:       u.ParentHash,
			SHA3Uncles:       u.SHA3Uncles,
			LogsBloom:        u.LogsBloom,
			TransactionsRoot: u.TransactionsRoot,
			StateRoot:        u.StateRoot,
			ReceiptsRoot:     u.ReceiptsRoot,
			Miner:            u.Miner,
			Author:           u.Author,
			Difficulty:       u.Difficulty,
			TotalDifficulty:  u.TotalDifficulty,
			ExtraData:        u.ExtraData,
			Size:             u.Size,
			GasLimit:         u.GasLimit,
			GasUsed:          u.GasUsed,
			Timestamp:        u.Timestamp,
			Transactions:     u.Transactions,
			Uncles:           u.Uncles,
			SealFields:       u.SealFields,
			Step:             u.Step,
			Signature:        u.Signature,
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
			TotalDifficulty  *Quantity  `json:"totalDifficulty"`
			ExtraData        Data       `json:"extraData"`
			Size             *Quantity  `json:"size"`
			GasLimit         Quantity   `json:"gasLimit"`
			GasUsed          Quantity   `json:"gasUsed"`
			Timestamp        Quantity   `json:"timestamp"`
			Transactions     []TxOrHash `json:"transactions"`
			Uncles           []Hash     `json:"uncles"`

			SealFields *[]Data `json:"sealFields,omitempty"`
		}

		c := clique{
			Number:           u.Number,
			Hash:             u.Hash,
			ParentHash:       u.ParentHash,
			SHA3Uncles:       u.SHA3Uncles,
			LogsBloom:        u.LogsBloom,
			TransactionsRoot: u.TransactionsRoot,
			StateRoot:        u.StateRoot,
			ReceiptsRoot:     u.ReceiptsRoot,
			Miner:            u.Miner,
			Author:           u.Author,
			Difficulty:       u.Difficulty,
			TotalDifficulty:  u.TotalDifficulty,
			ExtraData:        u.ExtraData,
			Size:             u.Size,
			GasLimit:         u.GasLimit,
			GasUsed:          u.GasUsed,
			Timestamp:        u.Timestamp,
			Transactions:     u.Transactions,
			Uncles:           u.Uncles,
			SealFields:       u.SealFields,
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
			TotalDifficulty  *Quantity  `json:"totalDifficulty"`
			ExtraData        Data       `json:"extraData"`
			Size             *Quantity  `json:"size"`
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
			Number:           u.Number,
			Hash:             u.Hash,
			ParentHash:       u.ParentHash,
			SHA3Uncles:       u.SHA3Uncles,
			LogsBloom:        u.LogsBloom,
			TransactionsRoot: u.TransactionsRoot,
			StateRoot:        u.StateRoot,
			ReceiptsRoot:     u.ReceiptsRoot,
			Miner:            u.Miner,
			Author:           u.Author,
			Difficulty:       u.Difficulty,
			TotalDifficulty:  u.TotalDifficulty,
			ExtraData:        u.ExtraData,
			Size:             u.Size,
			GasLimit:         u.GasLimit,
			GasUsed:          u.GasUsed,
			Timestamp:        u.Timestamp,
			Transactions:     u.Transactions,
			Uncles:           u.Uncles,
			Nonce:            u.Nonce,
			MixHash:          u.MixHash,
			SealFields:       u.SealFields,
		}

		return json.Marshal(&e)
	}

	type unknown Uncle
	unk := unknown(u)
	return json.Marshal(&unk)
}
