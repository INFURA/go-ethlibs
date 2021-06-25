package eth

import (
	"github.com/pkg/errors"

	"github.com/INFURA/go-ethlibs/rlp"
)

type AccessList []AccessListEntry

type AccessListEntry struct {
	Address     Address  `json:"address"`
	StorageKeys []Data32 `json:"storageKeys"`
}

// RLP returns the AccessList as an RLP-encoded list
func (a *AccessList) RLP() rlp.Value {
	if a == nil {
		// return empty list
		return rlp.Value{}
	}

	al := *a
	val := rlp.Value{List: make([]rlp.Value, len(al))}
	for i := range al {
		keys := rlp.Value{List: make([]rlp.Value, len(al[i].StorageKeys))}
		for j, k := range al[i].StorageKeys {
			keys.List[j] = k.RLP()
		}
		val.List[i] = rlp.Value{List: []rlp.Value{
			al[i].Address.RLP(),
			keys,
		}}
	}
	return val
}

// NewAccessListFromRLP decodes an RLP list into an AccessList, or returns an error.
// The RLP format of AccessLists is defined in EIP-2930, each entry is a tuple of an address and a list of storage slots.
func NewAccessListFromRLP(v rlp.Value) (AccessList, error) {
	accessList := make(AccessList, len(v.List))
	for j, accessRLP := range v.List {
		l := len(accessRLP.List)
		if l == 0 || l > 2 {
			return nil, errors.Errorf("invalid access list entry %d", j)
		}
		address, err := NewAddress(accessRLP.List[0].String)
		if err != nil {
			return nil, errors.Wrapf(err, "invalid access list entry address %d", j)
		}
		accessList[j].Address = *address
		if l == 2 {
			// 2nd item is the storage keys
			accessList[j].StorageKeys = make([]Data32, len(accessRLP.List[1].List))
			for k, key := range accessRLP.List[1].List {
				d, err := NewData32(key.String)
				if err != nil {
					return nil, errors.Wrapf(err, "invalid access list entry %d storage key %d", j, k)
				}
				accessList[j].StorageKeys[k] = *d
			}
		}
	}

	return accessList, nil
}
