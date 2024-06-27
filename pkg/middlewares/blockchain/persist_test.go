package blockchain_test

import (
	"bytes"
	"encoding/binary"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xoctopus/x/misc/must"

	. "github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/blockchain"
)

func dir(t *testing.T) string {
	path := filepath.Join(os.TempDir(), t.Name())
	must.NoError(os.RemoveAll(path))
	must.NoError(os.MkdirAll(path, 0777))
	return path
}

func TestPersist_Init(t *testing.T) {
	r := require.New(t)

	t.Run("InvalidPath", func(t *testing.T) {
		r.Error((&Persist{}).Init())
		r.True((&Persist{}).IsZero())
	})

	t.Run("Success", func(t *testing.T) {
		p := &Persist{Path: dir(t)}
		r.NoError(p.Init())
		r.NoError(p.Close())
	})
}

func TestPersist_Insert(t *testing.T) {
	r := require.New(t)

	p := &Persist{Path: dir(t)}
	r.NoError(p.Init())
	defer p.Close()

	meta := Meta{
		Network:  NETWORK__IOTX_TESTNET,
		Contract: common.HexToAddress("any"),
		Topic:    common.BytesToHash([]byte("any")),
	}
	logs := []*types.Log{
		{
			Address:     meta.Contract,
			Topics:      []common.Hash{meta.Topic},
			BlockNumber: 100,
			TxHash:      common.BytesToHash([]byte("logs1")),
		},
		{
			Address:     meta.Contract,
			Topics:      []common.Hash{meta.Topic},
			BlockNumber: 101,
			TxHash:      common.BytesToHash([]byte("logs2")),
		},
		{
			Address:     meta.Contract,
			Topics:      []common.Hash{meta.Topic},
			BlockNumber: 102,
			TxHash:      common.BytesToHash([]byte("logs2")),
		},
	}
	err := p.InsertLogs(meta, logs...)
	r.NoError(err)
	from, end, err := p.MetaRange(meta)
	r.NoError(err)
	r.Equal(from, uint64(0))
	r.Equal(end, uint64(102))

	lower := &types.Log{
		Address:     meta.Contract,
		Topics:      []common.Hash{meta.Topic},
		BlockNumber: 99,
		TxHash:      common.BytesToHash([]byte("lower")),
	}
	r.NoError(p.InsertLogs(meta, lower))
	from, end, err = p.MetaRange(meta)
	r.NoError(err)
	r.Equal(from, uint64(0))
	r.Equal(end, uint64(102))

	higher := &types.Log{
		Address:     meta.Contract,
		Topics:      []common.Hash{meta.Topic},
		BlockNumber: 150,
		TxHash:      common.BytesToHash([]byte("lower")),
	}
	r.NoError(p.InsertLogs(meta, higher))
	from, end, err = p.MetaRange(meta)
	r.NoError(err)
	r.Equal(from, uint64(0))
	r.Equal(end, higher.BlockNumber)
}

func TestPersist_GetSetUint64(t *testing.T) {
	r := require.New(t)
	p := &Persist{Path: dir(t)}
	r.NoError(p.Init())
	defer p.Close()

	key := []byte("any")
	t.Run("KeyNotFound", func(t *testing.T) {
		v, err := p.GetUint64(key)
		r.Equal(v, uint64(0))
		r.NoError(err)

		val, err := p.Load([]byte("other"))
		r.Nil(val)
		r.ErrorIs(err, pebble.ErrNotFound)
	})
	t.Run("UnexpectValueLength", func(t *testing.T) {
		val := make([]byte, 7)
		r.NoError(p.Store(key, val))
		defer p.Delete(key)

		v, err := p.GetUint64(key)
		r.Equal(v, uint64(0))
		r.ErrorContains(err, "expect uint64 value length is 8")
	})
	t.Run("Success", func(t *testing.T) {
		val1 := [8]byte{}
		binary.BigEndian.PutUint64(val1[:], 100)
		val2 := uint64(100)

		key1 := []byte("key1")
		r.NoError(p.Store(key1, val1[:]))

		key2 := []byte("key2")
		r.NoError(p.SetUint64(key2, val2))

		get1, err := p.Load(key1)
		r.NoError(err)
		get2, err := p.GetUint64(key2)
		r.NoError(err)

		r.Equal(binary.BigEndian.Uint64(get1), get2)
	})
}

func TestPersist_Queries(t *testing.T) {
	r := require.New(t)

	p := &Persist{Path: dir(t)}
	r.NoError(p.Init())
	defer p.Close()

	meta := Meta{
		Network:  NETWORK__IOTX_TESTNET,
		Contract: common.HexToAddress("any"),
		Topic:    common.BytesToHash([]byte("any")),
	}
	logs := []*types.Log{
		{
			Address:     meta.Contract,
			Topics:      []common.Hash{meta.Topic},
			BlockNumber: 100,
			TxHash:      common.BytesToHash([]byte("logs1")),
			Data:        []byte{},
		},
		{
			Address:     meta.Contract,
			Topics:      []common.Hash{meta.Topic},
			BlockNumber: 101,
			TxHash:      common.BytesToHash([]byte("logs2")),
			Data:        []byte{},
		},
		{
			Address:     meta.Contract,
			Topics:      []common.Hash{meta.Topic},
			BlockNumber: 102,
			TxHash:      common.BytesToHash([]byte("logs3")),
			Data:        []byte{},
		},
		{
			Address:     meta.Contract,
			Topics:      []common.Hash{meta.Topic},
			BlockNumber: 102,
			TxHash:      common.BytesToHash([]byte("logs4")),
			Data:        []byte{},
		},
	}

	err := p.InsertLogs(meta, logs...)
	r.NoError(err)

	t.Run("QueryTxByHash", func(t *testing.T) {
		l, err := p.QueryTxByHash(meta, logs[0].TxHash)
		r.NoError(err)
		r.Equal(*l, *logs[0])

		l, err = p.QueryTxByHash(meta, logs[1].TxHash)
		r.NoError(err)
		r.Equal(*l, *logs[1])

		l, err = p.QueryTxByHash(meta, logs[2].TxHash)
		r.NoError(err)
		r.Equal(*l, *logs[2])
		t.Run("NotFound", func(t *testing.T) {
			l, err = p.QueryTxByHash(meta, common.BytesToHash([]byte("not found")))
			r.ErrorIs(err, pebble.ErrNotFound)
		})
	})

	t.Run("QueryTxByHeight", func(t *testing.T) {
		_logs, err := p.QueryTxByHeight(meta, 102)
		r.NoError(err)
		r.Len(_logs, 2)
		r.Equal(*_logs[0], *logs[2])
		r.Equal(*_logs[1], *logs[3])
		t.Run("FailedToNewIter", func(t *testing.T) {
			mock.Patch(p.DB().NewIter, func(*pebble.IterOptions) (*pebble.Iterator, error) {
				t.Log("patched")
				return nil, errors.New(t.Name())
			})
			logs, err := p.QueryTxByHeight(meta, 102)
			r.Nil(logs)
			r.ErrorContains(err, t.Name())
		})
	})

	t.Run("QueryTxByHeightRange", func(t *testing.T) {
		_logs, err := p.QueryTxByHeightRange(meta, 0, 102)
		r.NoError(err)
		r.Len(_logs, 4)
		r.Equal(*_logs[0], *logs[0])
		r.Equal(*_logs[1], *logs[1])
		r.Equal(*_logs[2], *logs[2])
		r.Equal(*_logs[3], *logs[3])
	})

	t.Run("QueryByRange", func(t *testing.T) {
		db, err := pebble.Open("", &pebble.Options{FS: vfs.NewMem()})
		r.NoError(err)

		defer db.Close()

		bat := db.NewBatch()
		defer bat.Close()
		r.NoError(bat.Set([]byte("key1"), []byte("value1"), pebble.Sync))
		r.NoError(bat.Set([]byte("key2"), []byte("value2"), pebble.Sync))
		r.NoError(bat.Set([]byte("key3"), []byte("value3"), pebble.Sync))
		r.NoError(bat.Set([]byte("key4"), []byte("value4"), pebble.Sync))
		r.NoError(bat.Set([]byte("key5"), []byte("value5"), pebble.Sync))
		r.NoError(bat.Set([]byte("key6"), []byte("value6"), pebble.Sync))
		r.NoError(bat.Set([]byte("key7"), []byte("value7"), pebble.Sync))

		r.NoError(bat.Commit(pebble.Sync))

		iter, err := db.NewIter(&pebble.IterOptions{
			LowerBound: []byte("key3"),
			UpperBound: []byte("key5"),
		})
		r.NoError(err)
		defer iter.Close()

		for iter.First(); iter.Valid(); iter.Next() {
			t.Log(string(iter.Key()))
		}
		// expect output `key3` and `key4`
	})
}

func TestPersist_LoadAndDelete(t *testing.T) {
	r := require.New(t)

	p := &Persist{Path: dir(t)}
	r.NoError(p.Init())
	defer p.Close()

	key := []byte("key")

	val, loaded, err := p.LoadAndDelete(key)
	r.Nil(val)
	r.False(loaded)
	r.ErrorIs(err, pebble.ErrNotFound)

	r.NoError(p.Store(key, []byte("value")))
	val, loaded, err = p.LoadAndDelete(key)
	r.NoError(err)
	r.True(loaded)
	r.Equal(val, []byte("value"))

	_, err = p.Load(key)
	r.ErrorIs(err, pebble.ErrNotFound)
}

func TestPersist_LoadOrStore(t *testing.T) {
	r := require.New(t)

	p := &Persist{Path: dir(t)}
	r.NoError(p.Init())
	defer p.Close()

	key := []byte("key")

	t.Run("Unloaded", func(t *testing.T) {
		val, loaded, err := p.LoadOrStore(key, []byte("val"))
		r.NoError(err)
		r.False(loaded)
		r.Equal(val, []byte("val"))
	})

	t.Run("Loaded", func(t *testing.T) {
		val, loaded, err := p.LoadOrStore(key, []byte("other"))
		r.NoError(err)
		r.True(loaded)
		r.Equal(val, []byte("val"))
	})

	t.Run("FailedToGet", func(t *testing.T) {
		bat := p.DB().NewIndexedBatch()
		mock.Patch(p.DB().NewIndexedBatch, func() *pebble.Batch { return bat })
		mock.Patch(bat.Get, func([]byte) ([]byte, io.Closer, error) {
			return nil, nil, errors.New(t.Name())
		})
		_, _, err := p.LoadOrStore(key, []byte("any"))
		r.ErrorContains(err, t.Name())
	})
}

func TestPersist_BatchSet(t *testing.T) {
	r := require.New(t)

	p := &Persist{Path: dir(t)}
	r.NoError(p.Init())
	defer p.Close()

	kvs := [][2][]byte{
		{[]byte("TEST__1"), []byte("hello")},
		{[]byte("TEST__2"), []byte("pebble")},
		{[]byte("TEST__3"), []byte("!")},
	}
	t.Run("FailedToSet", func(t *testing.T) {
		bat := p.DB().NewIndexedBatch()
		mock.Patch(p.DB().NewIndexedBatch, func() *pebble.Batch { return bat })
		mock.Patch(bat.Set, func(k []byte, v []byte, _ *pebble.WriteOptions) error {
			if bytes.Equal(k, []byte("TEST__3")) {
				return errors.New(t.Name() + "::SET")
			}
			return nil
		})
		mock.Patch(bat.Commit, func(_ *pebble.WriteOptions) error {
			return errors.New(t.Name() + "::COMMIT")
		})
		err := p.BatchSet(kvs...)
		r.ErrorContains(err, "::SET")
		r.ErrorContains(err, "::COMMIT")
	})
	t.Run("Success", func(t *testing.T) {
		err := p.BatchSet(kvs...)
		r.NoError(err)
		keys, err := p.Keys([]byte("TEST__"))
		r.NoError(err)
		r.Len(keys, 3)
		for i, key := range keys {
			r.Equal(key, kvs[i][0])
		}
	})
}

func TestPersist_GetSetJSONTextValue(t *testing.T) {
	r := require.New(t)

	p := &Persist{Path: dir(t)}
	r.NoError(p.Init())
	defer p.Close()

	// big.Int implements JSON and text Marshaler/Unmarshaler
	v, _ := new(big.Int).SetString("999", 10)
	r.NoError(p.StoreJSONValue([]byte("json_big"), v))
	r.NoError(p.StoreTextValue([]byte("text_big"), v))

	vjson := new(big.Int)
	r.NoError(p.LoadJSONValue([]byte("json_big"), vjson))
	vtext := new(big.Int)
	r.NoError(p.LoadTextValue([]byte("text_big"), vtext))

	r.Equal(vjson.Cmp(v), 0)
	r.Equal(vjson.Cmp(vtext), 0)
}

func TestPersist_QueryAndUpdateWatcher(t *testing.T) {
	r := require.New(t)

	p := &Persist{Path: dir(t)}
	r.NoError(p.Init())
	defer p.Close()

	meta := Meta{
		Network:  NETWORK__IOTX_TESTNET,
		Contract: common.Address{},
		Topic:    common.Hash{},
	}
	from, end, err := p.QueryWatcher(meta, "sub")
	r.Equal(from, uint64(0))
	r.Equal(end, uint64(0))

	start, current := uint64(1000), uint64(1009)
	err = p.UpdateWatcher(meta, "sub", start, current)
	r.NoError(err)

	start2, current2, err := p.QueryWatcher(meta, "sub")
	r.NoError(err)
	r.Equal(start2, start)
	r.Equal(current2, current)
}
