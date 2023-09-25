package eth

import (
	"encoding/json"
	"strings"
)

type NewHeadsNotificationParams struct {
	Subscription string         `json:"subscription"`
	Result       NewHeadsResult `json:"result"`
}

// NewHeadsResult is the "result" payload in a newHeads notification.
// It looks a lot like a Block but is missing important fields, namely .TotalDifficulty and .Uncles
type NewHeadsResult struct {
	Number           Quantity `json:"number"`
	Hash             Hash     `json:"hash"`
	ParentHash       Hash     `json:"parentHash"`
	SHA3Uncles       Data32   `json:"sha3Uncles"`
	LogsBloom        Data256  `json:"logsBloom"`
	TransactionsRoot Data32   `json:"transactionsRoot"`
	StateRoot        Data32   `json:"stateRoot"`
	ReceiptsRoot     Data32   `json:"receiptsRoot"`
	Miner            Address  `json:"miner"`
	Author           Address  `json:"author,omitempty"` // Parity-specific alias of miner
	Difficulty       Quantity `json:"difficulty"`
	// TotalDifficulty  Quantity   `json:"totalDifficulty"`
	ExtraData Data      `json:"extraData"`
	Size      *Quantity `json:"size,omitempty"` // parity includes this geth does not
	GasLimit  Quantity  `json:"gasLimit"`
	GasUsed   Quantity  `json:"gasUsed"`
	Timestamp Quantity  `json:"timestamp"`
	// TODO: Support Transactions
	// Transactions     []TxOrHash `json:"transactions"`
	// Uncles           []Hash     `json:"uncles"`

	// EIP-1559 BaseFeePerGas
	BaseFeePerGas *Quantity `json:"baseFeePerGas,omitempty"`

	// EIP-4895 Withdrawals
	WithdrawalsRoot *Data32 `json:"withdrawalsRoot,omitempty"`

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

// FromBlock can be used to populate a NewHeadsResult with the contents of a Block.
// It does a best effort to emulate a NewHeadsResult with the same flavor as the Block.
func (nh *NewHeadsResult) FromBlock(block *Block) {

	*nh = NewHeadsResult{
		Number:           *block.Number,
		Hash:             *block.Hash,
		ParentHash:       block.ParentHash,
		SHA3Uncles:       block.SHA3Uncles,
		LogsBloom:        block.LogsBloom,
		TransactionsRoot: block.TransactionsRoot,
		StateRoot:        block.StateRoot,
		ReceiptsRoot:     block.ReceiptsRoot,
		Miner:            block.Miner,
		Author:           block.Author,
		Difficulty:       block.Difficulty,
		ExtraData:        block.ExtraData,
		// Size:          SEE BELOW,
		GasLimit:  block.GasLimit,
		GasUsed:   block.GasUsed,
		Timestamp: block.Timestamp,
		// Transactions:     nh.Transactions,
		Nonce:      block.Nonce,
		MixHash:    block.MixHash,
		SealFields: block.SealFields,
		Step:       block.Step,
		Signature:  block.Signature,

		// EIP-1559 BaseFeePerGas
		BaseFeePerGas: block.BaseFeePerGas,

		// EIP-4895 Withdrawals
		WithdrawalsRoot: block.WithdrawalsRoot,

		// EIP-4788 Beacon Block Root
		ParentBeaconBlockRoot: block.ParentBeaconBlockRoot,

		// EIP-4844 Blob related block fields
		ExcessBlobGas: block.ExcessBlobGas,
		BlobGasUsed:   block.BlobGasUsed,

		flavor: block.flavor,
	}

	// Parity includes .Size in its newHeads results while geth doesn't
	if strings.HasPrefix(nh.flavor, "parity") {
		size := block.Size
		nh.Size = &size
	}
}

func (nh *NewHeadsResult) UnmarshalJSON(data []byte) error {
	type alias NewHeadsResult
	aliased := alias(*nh)

	err := json.Unmarshal(data, &aliased)
	if err != nil {
		return err
	}

	*nh = NewHeadsResult(aliased)
	if nh.SealFields == nil {
		// It's a geth response, which is always the same regardless of consensus algorithm
		nh.flavor = "geth"
	} else {
		// It's a parity response, which differs by the consensus algorithm
		nh.flavor = "parity-unknown"

		if nh.MixHash != nil {
			// it's ethhash
			nh.flavor = "parity-ethhash"
		} else if nh.Step != nil && nh.Signature != nil {
			// it's Aura-based POA
			nh.flavor = "parity-aura"
		} else {
			// it's Clique-based POA
			nh.flavor = "parity-clique"
		}
	}
	return nil
}

func (nh NewHeadsResult) MarshalJSON() ([]byte, error) {
	switch nh.flavor {
	case "geth":
		type geth struct {
			Number           Quantity `json:"number"`
			Hash             Hash     `json:"hash"`
			ParentHash       Hash     `json:"parentHash"`
			SHA3Uncles       Data32   `json:"sha3Uncles"`
			LogsBloom        Data256  `json:"logsBloom"`
			TransactionsRoot Data32   `json:"transactionsRoot"`
			StateRoot        Data32   `json:"stateRoot"`
			ReceiptsRoot     Data32   `json:"receiptsRoot"`
			Miner            Address  `json:"miner"`
			Difficulty       Quantity `json:"difficulty"`
			// TotalDifficulty  Quantity   `json:"totalDifficulty"`
			ExtraData Data `json:"extraData"`
			// Size      Quantity `json:"size"` not included by geth
			GasLimit  Quantity `json:"gasLimit"`
			GasUsed   Quantity `json:"gasUsed"`
			Timestamp Quantity `json:"timestamp"`
			// Transactions     []TxOrHash `json:"transactions"`
			// Uncles           []Hash     `json:"uncles"`

			// EIP-1559 BaseFeePerGas
			BaseFeePerGas *Quantity `json:"baseFeePerGas,omitempty"`

			// EIP-4895 Withdrawals
			WithdrawalsRoot *Data32 `json:"withdrawalsRoot,omitempty"`

			// EIP-4788 Beacon Block Root
			ParentBeaconBlockRoot *Hash `json:"parentBeaconBlockRoot,omitempty"`

			// EIP-4844 Blob related block fields
			ExcessBlobGas *Quantity `json:"excessBlobGas,omitempty"`
			BlobGasUsed   *Quantity `json:"blobGasUsed,omitempty"`

			Nonce   *Data8 `json:"nonce"`
			MixHash *Data  `json:"mixHash"`
		}

		g := geth{
			Number:           nh.Number,
			Hash:             nh.Hash,
			ParentHash:       nh.ParentHash,
			SHA3Uncles:       nh.SHA3Uncles,
			LogsBloom:        nh.LogsBloom,
			TransactionsRoot: nh.TransactionsRoot,
			StateRoot:        nh.StateRoot,
			ReceiptsRoot:     nh.ReceiptsRoot,
			Miner:            nh.Miner,
			Difficulty:       nh.Difficulty,
			ExtraData:        nh.ExtraData,
			// Size:             nh.Size,
			GasLimit:  nh.GasLimit,
			GasUsed:   nh.GasUsed,
			Timestamp: nh.Timestamp,
			// Transactions:     nh.Transactions,
			BaseFeePerGas:         nh.BaseFeePerGas,
			WithdrawalsRoot:       nh.WithdrawalsRoot,
			ParentBeaconBlockRoot: nh.ParentBeaconBlockRoot,
			ExcessBlobGas:         nh.ExcessBlobGas,
			BlobGasUsed:           nh.BlobGasUsed,
			Nonce:                 nh.Nonce,
			MixHash:               nh.MixHash,
		}

		return json.Marshal(&g)
	case "parity-aura":
		type aura struct {
			Number           Quantity `json:"number"`
			Hash             Hash     `json:"hash"`
			ParentHash       Hash     `json:"parentHash"`
			SHA3Uncles       Data32   `json:"sha3Uncles"`
			LogsBloom        Data256  `json:"logsBloom"`
			TransactionsRoot Data32   `json:"transactionsRoot"`
			StateRoot        Data32   `json:"stateRoot"`
			ReceiptsRoot     Data32   `json:"receiptsRoot"`
			Miner            Address  `json:"miner"`
			Author           Address  `json:"author,omitempty"` // Parity-specific alias of miner
			Difficulty       Quantity `json:"difficulty"`
			// TotalDifficulty  Quantity   `json:"totalDifficulty"`
			ExtraData Data     `json:"extraData"`
			Size      Quantity `json:"size"`
			GasLimit  Quantity `json:"gasLimit"`
			GasUsed   Quantity `json:"gasUsed"`
			Timestamp Quantity `json:"timestamp"`
			// Transactions     []TxOrHash `json:"transactions"`
			// Uncles           []Hash     `json:"uncles"`

			SealFields *[]Data `json:"sealFields,omitempty"`
			Step       *string `json:"step,omitempty"`
			Signature  *string `json:"signature,omitempty"`
		}

		a := aura{
			Number:           nh.Number,
			Hash:             nh.Hash,
			ParentHash:       nh.ParentHash,
			SHA3Uncles:       nh.SHA3Uncles,
			LogsBloom:        nh.LogsBloom,
			TransactionsRoot: nh.TransactionsRoot,
			StateRoot:        nh.StateRoot,
			ReceiptsRoot:     nh.ReceiptsRoot,
			Miner:            nh.Miner,
			Author:           nh.Author,
			Difficulty:       nh.Difficulty,
			ExtraData:        nh.ExtraData,
			Size:             *nh.Size,
			GasLimit:         nh.GasLimit,
			GasUsed:          nh.GasUsed,
			Timestamp:        nh.Timestamp,
			// Transactions:     nh.Transactions,
			SealFields: nh.SealFields,
			Step:       nh.Step,
			Signature:  nh.Signature,
		}

		return json.Marshal(&a)
	case "parity-clique":
		type clique struct {
			Number           Quantity `json:"number"`
			Hash             Hash     `json:"hash"`
			ParentHash       Hash     `json:"parentHash"`
			SHA3Uncles       Data32   `json:"sha3Uncles"`
			LogsBloom        Data256  `json:"logsBloom"`
			TransactionsRoot Data32   `json:"transactionsRoot"`
			StateRoot        Data32   `json:"stateRoot"`
			ReceiptsRoot     Data32   `json:"receiptsRoot"`
			Miner            Address  `json:"miner"`
			Author           Address  `json:"author,omitempty"` // Parity-specific alias of miner
			Difficulty       Quantity `json:"difficulty"`
			// TotalDifficulty  Quantity   `json:"totalDifficulty"`
			ExtraData Data     `json:"extraData"`
			Size      Quantity `json:"size"`
			GasLimit  Quantity `json:"gasLimit"`
			GasUsed   Quantity `json:"gasUsed"`
			Timestamp Quantity `json:"timestamp"`
			// Transactions     []TxOrHash `json:"transactions"`
			// Uncles           []Hash     `json:"uncles"`

			SealFields *[]Data `json:"sealFields,omitempty"`
		}

		c := clique{
			Number:           nh.Number,
			Hash:             nh.Hash,
			ParentHash:       nh.ParentHash,
			SHA3Uncles:       nh.SHA3Uncles,
			LogsBloom:        nh.LogsBloom,
			TransactionsRoot: nh.TransactionsRoot,
			StateRoot:        nh.StateRoot,
			ReceiptsRoot:     nh.ReceiptsRoot,
			Miner:            nh.Miner,
			Author:           nh.Author,
			Difficulty:       nh.Difficulty,
			ExtraData:        nh.ExtraData,
			Size:             *nh.Size,
			GasLimit:         nh.GasLimit,
			GasUsed:          nh.GasUsed,
			Timestamp:        nh.Timestamp,
			// Transactions:     nh.Transactions,
			SealFields: nh.SealFields,
		}

		return json.Marshal(&c)
	case "parity-ethhash":
		type ethhash struct {
			Number           Quantity `json:"number"`
			Hash             Hash     `json:"hash"`
			ParentHash       Hash     `json:"parentHash"`
			SHA3Uncles       Data32   `json:"sha3Uncles"`
			LogsBloom        Data256  `json:"logsBloom"`
			TransactionsRoot Data32   `json:"transactionsRoot"`
			StateRoot        Data32   `json:"stateRoot"`
			ReceiptsRoot     Data32   `json:"receiptsRoot"`
			Miner            Address  `json:"miner"`
			Author           Address  `json:"author,omitempty"` // Parity-specific alias of miner
			Difficulty       Quantity `json:"difficulty"`
			// TotalDifficulty  Quantity   `json:"totalDifficulty"`
			ExtraData Data     `json:"extraData"`
			Size      Quantity `json:"size"`
			GasLimit  Quantity `json:"gasLimit"`
			GasUsed   Quantity `json:"gasUsed"`
			Timestamp Quantity `json:"timestamp"`
			// Transactions     []TxOrHash `json:"transactions"`
			// Uncles           []Hash     `json:"uncles"`

			Nonce      *Data8  `json:"nonce"`
			MixHash    *Data   `json:"mixHash"`
			SealFields *[]Data `json:"sealFields,omitempty"`
		}

		e := ethhash{
			Number:           nh.Number,
			Hash:             nh.Hash,
			ParentHash:       nh.ParentHash,
			SHA3Uncles:       nh.SHA3Uncles,
			LogsBloom:        nh.LogsBloom,
			TransactionsRoot: nh.TransactionsRoot,
			StateRoot:        nh.StateRoot,
			ReceiptsRoot:     nh.ReceiptsRoot,
			Miner:            nh.Miner,
			Author:           nh.Author,
			Difficulty:       nh.Difficulty,
			ExtraData:        nh.ExtraData,
			Size:             *nh.Size,
			GasLimit:         nh.GasLimit,
			GasUsed:          nh.GasUsed,
			Timestamp:        nh.Timestamp,
			// Transactions:     nh.Transactions,
			Nonce:      nh.Nonce,
			MixHash:    nh.MixHash,
			SealFields: nh.SealFields,
		}

		return json.Marshal(&e)
	}

	type unknown NewHeadsResult
	u := unknown(nh)
	return json.Marshal(&u)
}
