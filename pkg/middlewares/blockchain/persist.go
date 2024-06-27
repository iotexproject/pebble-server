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
	MetaRange(Meta) (uint64, uint64, error)
	// UpdateMetaRange update persisted meta range
	UpdateMetaRange(Meta, uint64, uint64) error
	// QueryTxByHash query tx log by tx hash and blockchain meta
	QueryTxByHash(Meta, common.Hash) (*types.Log, error)
	// QueryTxByHeightRange query tx logs by blockchain meta and block number range
	QueryTxByHeightRange(Meta, uint64, uint64) ([]*types.Log, error)
	// QueryTxByHeight query tx logs by blockchain meta and block number range
	QueryTxByHeight(Meta, uint64) ([]*types.Log, error)
	// InsertLogs insert tx logs with blockchain meta
	InsertLogs(Meta, ...*types.Log) error
	// QueryWatcher query watcher state by meta id and subscriber id
	QueryWatcher(Meta, string) (uint64, uint64, error)
	// UpdateWatcher update watcher state
	UpdateWatcher(Meta, string, uint64, uint64) error
	// RemoveWatcher remove watcher state
	RemoveWatcher(Meta, string) error
	// Close
	Close() error
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
		return errors.Wrapf(err, "path: %s", p.Path)
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
	defer func() {
		if recover() != nil {
			println(string(data))
		}
	}()
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

func (p *Persist) MetaRange(meta Meta) (from uint64, end uint64, err error) {
	from, err = p.GetUint64(meta.RangeFromKey())
	if err != nil {
		return
	}
	end, err = p.GetUint64(meta.RangeEndKey())
	if err != nil {
		return
	}
	return from, end, nil
}

func (p *Persist) UpdateMetaRange(meta Meta, from, end uint64) error {
	if err := p.SetUint64(meta.RangeFromKey(), from); err != nil {
		return err
	}
	return p.SetUint64(meta.RangeEndKey(), end)
}

// QueryTxByHeightRange query tx logs the block number between from and to [from,to]
func (p *Persist) QueryTxByHeightRange(meta Meta, from, to uint64) ([]*types.Log, error) {
	must.BeTrueWrap(from <= to, "assertion range start less than end")

	iter, err := p.db.NewIter(&pebble.IterOptions{
		LowerBound: meta.BlockKeyPrefixLowerBound(from),
		UpperBound: meta.BlockKeyPrefixLowerBound(to + 1),
	})
	if err != nil {
		return nil, err
	}

	logs := make([]*types.Log, 0)
	for iter.First(); iter.Valid(); iter.Next() {
		log := &types.Log{}
		key := iter.Key()
		if err = p.LoadJSONValue(key, log); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

func (p *Persist) QueryTxByHash(meta Meta, tx common.Hash) (*types.Log, error) {
	log := &types.Log{}
	if err := p.LoadJSONValue(meta.TxHashKey(tx), log); err != nil {
		return nil, err
	}
	return log, nil
}

func (p *Persist) QueryTxByHeight(meta Meta, blk uint64) ([]*types.Log, error) {
	iter, err := p.db.NewIter(&pebble.IterOptions{
		LowerBound: meta.BlockKeyPrefixLowerBound(blk),
		UpperBound: meta.BlockKeyPrefixUpperBound(blk),
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	logs := make([]*types.Log, 0)
	for iter.First(); iter.Valid(); iter.Next() {
		log := &types.Log{}
		key := iter.Key()
		if err = p.LoadJSONValue(key, log); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func (p *Persist) InsertLogs(meta Meta, logs ...*types.Log) error {
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
			[2][]byte{meta.TxHashKey(log.TxHash), data},
			// BLK_$(META)$(BLK)$(TX): LOG DATA
			[2][]byte{meta.BlockKey(log), data},
		)

		if log.BlockNumber > end {
			highest = log.BlockNumber
			end = highest
		}
	}

	if highest != 0 {
		blk := [8]byte{}
		gByteOrder.PutUint64(blk[:], highest)
		kvs = append(kvs, [2][]byte{meta.RangeEndKey(), blk[:]})
	}

	return p.BatchSet(kvs...)
}

func (p *Persist) QueryWatcher(meta Meta, id string) (uint64, uint64, error) {
	from, err := p.GetUint64(meta.WatcherFromKey(id))
	if err != nil {
		return 0, 0, err
	}
	end, err := p.GetUint64(meta.WatcherEndKey(id))
	if err != nil {
		return 0, 0, err
	}
	return from, end, nil
}

func (p *Persist) UpdateWatcher(meta Meta, id string, from, end uint64) error {
	if err := p.SetUint64(meta.WatcherFromKey(id), from); err != nil {
		return err
	}
	return p.SetUint64(meta.WatcherEndKey(id), end)
}

func (p *Persist) RemoveWatcher(meta Meta, id string) error {
	if err := p.Delete(meta.WatcherFromKey(id)); err != nil {
		return err
	}
	return p.Delete(meta.WatcherEndKey(id))
}

var gByteOrder = binary.BigEndian
