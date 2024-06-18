package blockchain

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"io"

	"github.com/cockroachdb/pebble"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/textx"
)

type TxPersistence interface {
	// MetaRange query persisted meta range
	MetaRange(MetaID) (uint64, uint64, error)
	// UpdateMetaRange update persisted meta range
	UpdateMetaRange(MetaID, uint64, uint64) error
	// QueryTxByHash query tx log by tx hash and blockchain meta
	QueryTxByHash(MetaID, common.Hash) (*types.Log, error)
	// QueryTxByHeightRange query tx logs by blockchain meta and block number range
	QueryTxByHeightRange(MetaID, uint64, uint64) ([]*types.Log, error)
	// QueryTxByHeight query tx logs by blockchain meta and block number range
	QueryTxByHeight(MetaID, uint64) ([]*types.Log, error)
	// InsertLogs insert tx logs with blockchain meta
	InsertLogs(MetaID, ...*types.Log) error
	// QueryWatcher query watcher state by meta id and subscriber id
	QueryWatcher(MetaID, string) (uint64, uint64, error)
	// UpdateWatcher update watcher state
	UpdateWatcher(MetaID, string, uint64, uint64) error
}

type PebbleKVStore interface {
	Load(k []byte) ([]byte, error)
	Store(k []byte, v []byte) error
	LoadAndDelete([]byte) (value []byte, loaded bool, err error)
	LoadOrStore([]byte, []byte) (actual []byte, loaded bool, err error)
	Delete([]byte) error
	Keys(prefix []byte) (keys [][]byte, err error)
}

var _ TxPersistence = (*Persist)(nil)

type Persist struct {
	Path string

	db *pebble.DB `env:"-"`
}

func (p *Persist) IsZero() bool {
	return p == nil || p.Path == ""
}

func (p *Persist) Init() error {
	db, err := pebble.Open(p.Path, &pebble.Options{})
	if err != nil {
		return err
	}
	p.db = db
	return nil
}

func (p *Persist) DB() *pebble.DB { return p.db }

func (p *Persist) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func (p *Persist) BatchSet(kvs ...[2][]byte) (err error) {
	bat := p.db.NewIndexedBatch()
	stk := &stacks{}

	defer func() {
		stk.Append(bat.Commit(pebble.Sync), "failed to commit")
		_ = bat.Close()
		err = stk.Final()
	}()

	for _, kv := range kvs {
		if err = bat.Set(kv[0], kv[1], pebble.Sync); err != nil {
			stk.Append(err, "failed to set %s", hex.EncodeToString(kv[0]))
			return
		}
	}
	return nil
}

func (p *Persist) Load(k []byte) ([]byte, error) {
	v, closer, err := p.db.Get(k)
	if err == nil {
		_ = closer.Close()
		return v, nil
	}
	return nil, err
}

func (p *Persist) Store(k, v []byte) error {
	return p.db.Set(k, v, pebble.Sync)
}

func (p *Persist) LoadAndDelete(k []byte) (v []byte, loaded bool, err error) {
	bat := p.db.NewIndexedBatch()
	stack := &stacks{}

	defer func() {
		err = stack.Final()
	}()

	defer func() {
		stack.Append(bat.Delete(k, pebble.Sync), "failed to delete")
		stack.Append(bat.Commit(pebble.Sync), "failed to commit")
		_ = bat.Close()
	}()

	var closer io.Closer

	v, closer, err = bat.Get(k)
	if err == nil {
		_ = closer.Close()
		loaded = true
		return
	}
	stack.Append(err, "failed to get")
	return
}

func (p *Persist) LoadOrStore(k, v []byte) (actual []byte, loaded bool, err error) {
	bat := p.db.NewIndexedBatch()
	stk := &stacks{}

	defer func() {
		stk.Append(bat.Commit(pebble.Sync), "failed to commit")
		_ = bat.Close()
		err = stk.Final()
		if err != nil {
			actual = nil
			return
		}
		if !loaded {
			actual = v
		}
	}()

	var closer io.Closer
	actual, closer, err = bat.Get(k)
	if err == nil {
		_ = closer.Close()
		loaded = true
		return
	}
	stk.Append(err, "failed to get")

	if errors.Is(err, pebble.ErrNotFound) {
		stk.TrimLast()
		stk.Append(bat.Set(k, v, pebble.Sync), "failed to set")
	}
	return
}

func (p *Persist) Delete(k []byte) error {
	return p.db.Delete(k, pebble.Sync)
}

func (p *Persist) Keys(prefix []byte) ([][]byte, error) {
	lowerBound := prefix
	upperbound := func(b []byte) []byte {
		end := make([]byte, len(b))
		copy(end, b)
		for i := len(end) - 1; i >= 0; i-- {
			end[i] = end[i] + 1
			if end[i] != 0 {
				return end[:i+1]
			}
		}
		return nil // no upper-bound
	}(lowerBound)
	iter, err := p.db.NewIter(&pebble.IterOptions{
		LowerBound: lowerBound,
		UpperBound: upperbound,
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	keys := make([][]byte, 0)
	for iter.First(); iter.Valid(); iter.Next() {
		key := make([]byte, len(iter.Key()))
		copy(key, iter.Key())
		keys = append(keys, key)
	}
	return keys, nil
}

func (p *Persist) GetUint64(k []byte) (uint64, error) {
	value, closer, err := p.db.Get(k)
	if err == nil {
		_ = closer.Close()

		if len(value) != 8 {
			return 0, errors.Errorf(
				"expect uint64 value length is 8 but got %d key: %s",
				len(value), hex.EncodeToString(k),
			)
		}
		v := gByteOrder.Uint64(value)
		return v, nil
	}
	if errors.Is(err, pebble.ErrNotFound) {
		return 0, nil
	}
	return 0, err
}

func (p *Persist) SetUint64(k []byte, v uint64) error {
	value := [8]byte{}
	gByteOrder.PutUint64(value[:], v)
	return p.db.Set(k, value[:], pebble.Sync)
}

func (p *Persist) StoreJSONValue(k []byte, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return p.Store(k, data)
}

func (p *Persist) LoadJSONValue(k []byte, v any) error {
	data, err := p.Load(k)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (p *Persist) StoreTextValue(k []byte, v any) error {
	data, err := textx.MarshalText(v)
	if err != nil {
		return err
	}
	return p.Store(k, data)
}

func (p *Persist) LoadTextValue(k []byte, v any) error {
	data, err := p.Load(k)
	if err != nil {
		return err
	}
	return textx.UnmarshalText(data, v)
}

func (p *Persist) MetaRange(meta MetaID) (from uint64, end uint64, err error) {
	from, err = p.GetUint64(MetaRangeFromKey(meta))
	if err != nil {
		return
	}
	end, err = p.GetUint64(MetaRangeEndKey(meta))
	if err != nil {
		return
	}
	return from, end, nil
}

func (p *Persist) UpdateMetaRange(meta MetaID, from, end uint64) error {
	if err := p.SetUint64(MetaRangeFromKey(meta), from); err != nil {
		return err
	}
	return p.SetUint64(MetaRangeEndKey(meta), end)
}

func (p *Persist) QueryTxByHeightRange(meta MetaID, from, to uint64) ([]*types.Log, error) {
	must.BeTrueWrap(from <= to, "assertion range start less than end")
	logs := make([]*types.Log, 0)
	for i := from; i <= to; i++ {
		_logs, err := p.QueryTxByHeight(meta, i)
		if err != nil {
			return nil, err
		}
		logs = append(logs, _logs...)
	}
	return logs, nil
}

func (p *Persist) QueryTxByHash(meta MetaID, tx common.Hash) (*types.Log, error) {
	log := &types.Log{}
	if err := p.LoadJSONValue(TxHashKey(meta, tx), log); err != nil {
		return nil, err
	}
	return log, nil
}

func (p *Persist) QueryTxByHeight(meta MetaID, blk uint64) ([]*types.Log, error) {
	lowerBound := BlockKeyPrefix(meta, blk)
	upperbound := func(b []byte) []byte {
		end := make([]byte, len(b))
		copy(end, b)
		for i := len(end) - 1; i >= 0; i-- {
			end[i] = end[i] + 1
			if end[i] != 0 {
				return end[:i+1]
			}
		}
		return nil // no upper-bound
	}(lowerBound)
	iter, err := p.db.NewIter(&pebble.IterOptions{
		LowerBound: lowerBound,
		UpperBound: upperbound},
	)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	logs := make([]*types.Log, 0)
	for iter.First(); iter.Valid(); iter.Next() {
		log := &types.Log{}
		if err = p.LoadJSONValue(iter.Key(), log); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func (p *Persist) InsertLogs(meta MetaID, logs ...*types.Log) error {
	_, end, err := p.MetaRange(meta)
	if err != nil {
		return err
	}

	var highest uint64

	var kvs [][2][]byte
	for i := range logs {
		log := logs[i]
		data, err := log.MarshalJSON()
		if err != nil {
			return errors.Wrapf(
				err, "failed to serialize log %s:%s",
				meta.String(), log.TxHash.String(),
			)
		}
		kvs = append(kvs,
			// LOG_$(META)$(TX): LOG DATA
			[2][]byte{TxHashKey(meta, log.TxHash), data},
			// BLK_$(META)$(BLK)$(TX): LOG DATA
			[2][]byte{BlockKey(meta, log.TxHash, log.BlockNumber), data},
		)

		if log.BlockNumber > end {
			highest = log.BlockNumber
			end = highest
		}
	}

	if highest != 0 {
		blk := [8]byte{}
		gByteOrder.PutUint64(blk[:], highest)
		kvs = append(kvs, [2][]byte{MetaRangeEndKey(meta), blk[:]})
	}

	return p.BatchSet(kvs...)
}

func (p *Persist) QueryWatcher(meta MetaID, id string) (uint64, uint64, error) {
	key := MetaSubKey(meta, id)
	val, err := p.Load(key)
	if err != nil {
		return 0, 0, err
	}
	if len(val) != 16 {
		return 0, 0, errors.Errorf(
			"expect meta sub key value length is 16 but got %d key: %s",
			len(val), hex.EncodeToString(key),
		)
	}
	return gByteOrder.Uint64(val[0:8]), gByteOrder.Uint64(val[8:16]), nil
}

func (p *Persist) UpdateWatcher(meta MetaID, id string, start, current uint64) error {
	key := MetaSubKey(meta, id)
	val := [16]byte{}
	gByteOrder.PutUint64(val[0:8], start)
	gByteOrder.PutUint64(val[8:16], current)
	return p.Store(key, val[:])
}

func TxHashKey(meta MetaID, tx common.Hash) []byte {
	key := make([]byte, HashKeyLength)
	offset := 0
	copy(key, _LOG)
	offset += 4
	copy(key[offset:], meta.Bytes())
	offset += MetaIDLength
	copy(key[offset:], tx.Bytes())
	offset += common.HashLength

	must.BeTrue(offset == HashKeyLength)
	return key
}

func BlockKey(meta MetaID, tx common.Hash, blk uint64) []byte {
	key := make([]byte, BlockKeyLength)
	offset := 0
	copy(key, _BLK)
	offset += 4
	copy(key[offset:], meta.Bytes())
	offset += MetaIDLength
	gByteOrder.PutUint64(key[offset:], blk)
	offset += 8
	copy(key[offset:], tx.Bytes())
	offset += common.HashLength

	must.BeTrue(offset == BlockKeyLength)
	return key
}

func BlockKeyPrefix(meta MetaID, blk uint64) []byte {
	key := make([]byte, BlockKeyPrefixLength)
	offset := 0
	copy(key, _BLK)
	offset += 4
	copy(key[offset:], meta.Bytes())
	offset += MetaIDLength
	gByteOrder.PutUint64(key[offset:], blk)
	offset += 8

	must.BeTrue(offset == BlockKeyPrefixLength)
	return key
}

func MetaRangeFromKey(meta MetaID) []byte {
	key := make([]byte, MetaRangeKeyLength)
	offset := 0
	copy(key, _RL)
	offset += 4
	copy(key[offset:], meta.Bytes())
	offset += MetaIDLength

	must.BeTrue(offset == MetaRangeKeyLength)
	return key
}

func MetaRangeEndKey(meta MetaID) []byte {
	key := make([]byte, MetaRangeKeyLength)
	offset := 0
	copy(key, _RH)
	offset += 4
	copy(key[offset:], meta.Bytes())
	offset += MetaIDLength

	must.BeTrue(offset == MetaRangeKeyLength)
	return key
}

func MetaSubKey(meta MetaID, sub string) []byte {
	length := len(_SUB) + MetaIDLength + len(sub)
	key := make([]byte, length)
	offset := 0
	copy(key, _SUB)
	offset += 4
	copy(key[offset:], meta.Bytes())
	offset += MetaIDLength
	copy(key[offset:], sub)
	offset += len(sub)

	must.BeTrue(offset == length)
	return key
}

var (
	_LOG = []byte("LOG_")
	_BLK = []byte("BLK_")
	_RL  = []byte("RNL_")
	_RH  = []byte("RNH_")
	_SUB = []byte("SUB_")

	gByteOrder = binary.BigEndian
)

const (
	HashKeyLength        = 4 + MetaIDLength + common.HashLength
	BlockKeyLength       = 4 + MetaIDLength + 8 + common.HashLength
	BlockKeyPrefixLength = 4 + MetaIDLength + 8
	MetaRangeKeyLength   = 4 + MetaIDLength
)
