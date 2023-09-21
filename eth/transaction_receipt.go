package eth

type TransactionReceipt struct {
	Type              *Quantity `json:"type,omitempty"`
	TransactionHash   Hash      `json:"transactionHash"`
	TransactionIndex  Quantity  `json:"transactionIndex"`
	BlockHash         Hash      `json:"blockHash"`
	BlockNumber       Quantity  `json:"blockNumber"`
	From              Address   `json:"from"`
	To                *Address  `json:"to"`
	CumulativeGasUsed Quantity  `json:"cumulativeGasUsed"`
	GasUsed           Quantity  `json:"gasUsed"`
	ContractAddress   *Address  `json:"contractAddress"`
	Logs              []Log     `json:"logs"`
	LogsBloom         Data256   `json:"logsBloom"`
	Root              *Data32   `json:"root,omitempty"`
	Status            *Quantity `json:"status,omitempty"`
	EffectiveGasPrice *Quantity `json:"effectiveGasPrice,omitempty"`

	// EIP-4844 Receipt Fields
	BlobGasPrice *Quantity `json:"blobGasPrice,omitempty"`
	BlobGasUsed  *Quantity `json:"blobGasUsed,omitempty"`
}

// TransactionType returns the transactions EIP-2718 type, or TransactionTypeLegacy for pre-EIP-2718 transactions.
func (t *TransactionReceipt) TransactionType() int64 {
	if t.Type == nil {
		return TransactionTypeLegacy
	}

	return t.Type.Int64()
}
