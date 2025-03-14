package eth

import (
	"github.com/INFURA/go-ethlibs/rlp"
	"github.com/pkg/errors"
)

type AuthorizationList []SetCodeAuthorization

type SetCodeAuthorization struct {
	ChainID *Quantity `json:"chainId"`
	Address Address   `json:"address"`
	Nonce   Quantity  `json:"nonce"`
	V       Quantity  `json:"yParity"`
	R       Quantity  `json:"r"`
	S       Quantity  `json:"s"`
}

func (a *AuthorizationList) RLP() rlp.Value {
	if a == nil {
		return rlp.Value{}
	}

	al := *a
	val := rlp.Value{List: make([]rlp.Value, len(al))}
	for i := range al {
		val.List[i] = rlp.Value{List: []rlp.Value{
			al[i].ChainID.RLP(),
			al[i].Address.RLP(),
			al[i].Nonce.RLP(),
			al[i].V.RLP(),
			al[i].R.RLP(),
			al[i].S.RLP(),
		}}
	}
	return val
}

// NewAuthorizationListFromRLP decodes an RLP list into an AuthorizationList, or returns an error.
// The RLP format of AuthorizationLists is defined in EIP-7702, each entry is a list of a chain ID, an address,
// and a list of authorization parameters.
func NewAuthorizationListFromRLP(v rlp.Value) (AuthorizationList, error) {
	al := make(AuthorizationList, len(v.List))

	for i := range v.List {
		authorization := v.List[i].List
		if len(authorization) < 5 {
			return nil, errors.New("invalid authorization")
		}

		chainID, err := NewQuantityFromRLP(authorization[0])
		if err != nil {
			return nil, errors.Wrapf(err, "invalid authorization %d chain ID", i)
		}
		address, err := NewAddress(authorization[1].String)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid authorization %d address", i)
		}
		nonce, err := NewQuantityFromRLP(authorization[2])
		if err != nil {
			return nil, errors.Wrapf(err, "invalid authorization %d nonce", i)
		}
		v, err := NewQuantityFromRLP(authorization[3])
		if err != nil {
			return nil, errors.Wrapf(err, "invalid authorization %d V", i)
		}
		r, err := NewQuantityFromRLP(authorization[4])
		if err != nil {
			return nil, errors.Wrapf(err, "invalid authorization %d R", i)
		}
		s, err := NewQuantityFromRLP(authorization[5])
		if err != nil {
			return nil, errors.Wrapf(err, "invalid authorization %d S", i)
		}

		al[i] = SetCodeAuthorization{
			ChainID: chainID,
			Address: *address,
			Nonce:   *nonce,
			V:       *v,
			R:       *r,
			S:       *s,
		}
	}
	return al, nil
}
