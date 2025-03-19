package eth

import "github.com/justinwongcn/go-ethlibs/rlp"

type Withdrawal struct {
	Index          Quantity `json:"index"`
	ValidatorIndex Quantity `json:"validatorIndex"`
	Address        Address  `json:"address"`
	Amount         Quantity `json:"amount"`
}

// RLP returns the EIP-4895 RLP encoding of a Withdrawal item
func (w *Withdrawal) RLP() rlp.Value {
	// withdrawal_0 = [index_0, validator_index_0, address_0, amount_0]
	return rlp.Value{
		List: []rlp.Value{
			w.Index.RLP(),
			w.ValidatorIndex.RLP(),
			w.Address.RLP(),
			w.Amount.RLP(),
		},
	}
}
