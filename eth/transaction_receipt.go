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
}
