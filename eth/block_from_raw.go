package eth

import (
	"github.com/pkg/errors"

	"github.com/INFURA/go-ethlibs/rlp"
)

// FromRaw populates Block fields from the RLP-encoded raw block input string.
func (b *Block) FromRaw(input string) error {
	// Decode the input string as an rlp.Value
	decoded, err := rlp.From(input)
	if err != nil {
		return errors.Wrap(err, "could not RLP decode raw input")
	}

	// decoded should be 3 lists: header, transactions, uncles
	switch len(decoded.List) {
	case 0:
		return errors.New("raw input decoded to non-list or empty list")
	case 3:
		// expected
		break
	default:
		return errors.Errorf("unexpected decoded list size %d", len(decoded.List))
	}

	// compute block header hash
	h, err := decoded.List[0].Hash()
	if err != nil {
		return errors.Wrap(err, "could not compute RLP hash")
	}
	hash, err := NewHash(h)

	header, txs, uncles := decoded.List[0].List, decoded.List[1].List, decoded.List[2].List
	// header should be 15 items for legacy blocks, 16 for EIP-1559 blocks
	switch len(header) {
	case 15, 16:
	default:
		return errors.Errorf("unexpected decoded header list size %d", len(header))
	}

	transactions := make([]TxOrHash, len(txs))
	for i, txRlp := range txs {
		index := QuantityFromInt64(int64(i))
		tx := Transaction{
			BlockHash: hash,
			Index:     &index,
		}
		// Each transaction in the txs RLP list is either an opaque binary blob (an EIP-2718 tx) or itself an RLP list
		// (a legacy pre-2718 transaction).  2718 transactions can be passed to Transaction.FromRaw as is, legacy
		// transactions need to be converted to raw blobs via rlp.Value.Encode first.
		rawTx := txRlp.String
		if txRlp.IsList() {
			rawTx, err = txRlp.Encode()
			if err != nil {
				return errors.Wrap(err, "could not re-encode transaction")
			}
		}
		if err := tx.FromRaw(rawTx); err != nil {
			return errors.Wrap(err, "could not decode transaction")
		}

		transactions[i] = TxOrHash{
			Transaction: tx,
			Populated:   true,
		}
	}

	uncleHashes := make([]Hash, len(uncles))
	for i, u := range uncles {
		h, err := u.Hash()
		if err != nil {
			return errors.Wrap(err, "could not hash uncle RLP data")
		}
		if hh, err := NewHash(h); err == nil {
			uncleHashes[i] = *hh
		} else {
			return errors.Wrap(err, "could not encode uncle to hash")
		}
	}

	// ParentHash
	if p, err := NewHash(header[0].String); err == nil {
		b.ParentHash = *p
	} else {
		return errors.Wrap(err, "could not convert header field 0 to ParentHash")
	}

	// SHA3Uncles
	if u, err := NewHash(header[1].String); err == nil {
		b.SHA3Uncles = *u
	} else {
		return errors.Wrap(err, "could not convert header field 1 to SHA3Uncles")
	}

	// Miner
	if m, err := NewAddress(header[2].String); err == nil {
		b.Miner = *m
	} else {
		return errors.Wrap(err, "could not convert header field 2 to Miner")
	}

	// StateRoot
	if s, err := NewData32(header[3].String); err == nil {
		b.StateRoot = *s
	} else {
		return errors.Wrap(err, "could not convert header field 3 to StateRoot")
	}

	// TransactionsRoot
	if t, err := NewData32(header[4].String); err == nil {
		b.TransactionsRoot = *t
	} else {
		return errors.Wrap(err, "could not convert header field 4 to TransactionsRoot")
	}

	// ReceiptsRoot
	if r, err := NewData32(header[5].String); err == nil {
		b.ReceiptsRoot = *r
	} else {
		return errors.Wrap(err, "could not convert header field 5 to ReceiptsRoot")
	}

	// LogsBloom
	if l, err := NewData256(header[6].String); err == nil {
		b.LogsBloom = *l
	} else {
		return errors.Wrap(err, "could not convert header field 6 to LogsBloom")
	}

	// Difficulty
	if q, err := NewQuantityFromRLP(header[7]); err == nil {
		b.Difficulty = *q
	} else {
		return errors.Wrap(err, "could not convert header field 7 to Difficulty")
	}

	// Number
	if q, err := NewQuantityFromRLP(header[8]); err == nil {
		b.Number = q
	} else {

		return errors.Wrap(err, "could not convert header field 8 to Number")
	}

	for i := range transactions {
		transactions[i].Transaction.BlockNumber = b.Number
	}

	// GasLimit
	if q, err := NewQuantityFromRLP(header[9]); err == nil {
		b.GasLimit = *q
	} else {
		return errors.Wrap(err, "could not convert header field 9 to GasLimit")
	}

	// GasUsed
	if q, err := NewQuantityFromRLP(header[10]); err == nil {
		b.GasUsed = *q
	} else {
		return errors.Wrap(err, "could not convert header field 10 to GasUsed")
	}

	// Timestamp
	if q, err := NewQuantityFromRLP(header[11]); err == nil {
		b.Timestamp = *q
	} else {
		return errors.Wrap(err, "could not convert header field 11 to Timestamp")
	}

	// ExtraData
	if d, err := NewData(header[12].String); err == nil {
		b.ExtraData = *d
	} else {
		return errors.Wrap(err, "could not convert header field 12 to ExtraData")
	}

	// MixHash
	if d, err := NewData(header[13].String); err == nil {
		b.MixHash = d
	} else {
		return errors.Wrap(err, "could not convert header field 13 to MixHash")
	}

	// Nonce
	if n, err := NewData8(header[14].String); err == nil {
		b.Nonce = n
	} else {
		return errors.Wrap(err, "could not convert header field 14 to Nonce")
	}

	// BaseFee (EIP-1559 enabled blocks)
	if len(header) >= 16 {
		q, err := NewQuantityFromRLP(header[15])
		if err != nil {
			return errors.Wrap(err, "could not convert header field 15 to BaseFeePerGas")
		}
		b.BaseFeePerGas = q
	}

	b.Hash = hash
	b.Uncles = uncleHashes
	b.Transactions = transactions
	return nil
}
