package eth

type GetLogsParameters struct {
	FromBlock *BlockNumber `json:"fromBlock,omitempty"`
	ToBlock   *BlockNumber `json:"tomBlock,omitempty"`
	BlockHash *Hash        `json:"blockHash,omitempty"`
	Address   *Address     `json:"address,omitempty"`
	Topics    []Topic      `json:"topics,omitempty"`
}
