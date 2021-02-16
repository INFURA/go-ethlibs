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

func (a AccessList) RLP() rlp.Value {
	val := rlp.Value{List: make([]rlp.Value, len(a))}
	for i := range a {
		keys := rlp.Value{List: make([]rlp.Value, len(a[i].StorageKeys))}
		for j, k := range a[i].StorageKeys {
			keys.List[j] = k.RLP()
		}
		val.List[i] = rlp.Value{List: []rlp.Value{
			a[i].Address.RLP(),
			keys,
		}}
	}
	return val
}

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
