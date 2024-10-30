package blockchain_test

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/xhd2015/xgo/runtime/mock"

	. "github.com/iotexproject/pebble-server/middlewares/blockchain"
)

func TestMonitor_Init(t *testing.T) {
	r := require.New(t)

	network := NETWORK__IOTX_TESTNET

	client := &EthClient{
		Endpoint: "https://babel-api.testnet.iotex.io",
		Network:  network,
	}
	r.NoError(client.Init())

	contract := &Contract{
		ID:      "any",
		Network: network,
		Address: common.HexToAddress("0x6AfCB0EB71B7246A68Bb9c0bFbe5cD7c11c4839f"),
		Events:  []*Event{{Name: "ProjectConfigUpdated", ABI: ProjectConfigUpdatedABI}},
	}
	r.NoError(contract.Init())

	meta := NewMeta(network, contract)

	t.Run("FailedToLoadMonitorRange", func(t *testing.T) {
		p := &Persist{Path: dir(t)}
		r.NoError(p.Init())
		defer p.Close()

		m := NewMonitor(network, contract).
			WithPersistence(p).
			WithEthClient(client)
		r.NoError(p.Store(meta.MonitorRangeEndKey(), make([]byte, 10)))
		err := m.Init()
		r.ErrorContains(err, "failed to load monitor range")
		r.Equal(m.Endpoint(), client.Endpoint)
	})

	t.Run("FromGenesisBlock", func(t *testing.T) {
		t.Run("WithoutStartBlock", func(t *testing.T) {
			p := &Persist{Path: dir(t)}
			r.NoError(p.Init())
			defer p.Close()

			m := NewMonitor(network, contract).
				WithPersistence(p).
				WithEthClient(client)

			err := m.Init()
			time.Sleep(5 * time.Millisecond)
			m.Stop()
			r.NoError(err)
			r.Equal(m.StartAt(), uint64(1))
			r.Equal(m.CurrentBlock(), uint64(1))
		})
		t.Run("WithStartBlock", func(t *testing.T) {
			p := &Persist{Path: dir(t)}
			r.NoError(p.Init())
			defer p.Close()
			m := NewMonitor(network, contract).
				WithPersistence(p).
				WithInterval(time.Second).
				WithEthClient(client).
				WithStartBlock(10000)
			err := m.Init()
			time.Sleep(5 * time.Millisecond)
			m.Stop()
			r.NoError(err)
			r.Equal(m.StartAt(), uint64(10000))
			r.Equal(m.CurrentBlock(), uint64(10000))
		})
	})

	t.Run("LoadOffsetFromPersistenceRecord", func(t *testing.T) {
		p := &Persist{Path: dir(t)}
		r.NoError(p.Init())
		defer p.Close()
		r.NoError(p.SetUint64(meta.MonitorRangeFromKey(), 1000))
		r.NoError(p.SetUint64(meta.MonitorRangeEndKey(), 2000))

		m := NewMonitor(network, contract).
			WithPersistence(p).
			WithEthClient(client).
			WithInterval(time.Second).
			WithStartBlock(10000)
		err := m.Init()
		time.Sleep(5 * time.Millisecond)
		m.Stop()
		r.NoError(err)
		r.Equal(m.StartAt(), uint64(1000))
		r.Equal(m.CurrentBlock(), uint64(2000))
	})

	t.Run("FailedToUpdateMetaRange", func(t *testing.T) {
		p := &Persist{Path: dir(t)}
		r.NoError(p.Init())
		defer p.Close()

		m := NewMonitor(network, contract).
			WithPersistence(p).
			WithEthClient(client).
			WithInterval(time.Second).
			WithStartBlock(10000)
		mock.Patch(p.UpdateMonitorRange, func(Meta, uint64, uint64) error {
			return errors.New(t.Name())
		})
		err := m.Init()
		r.ErrorContains(err, "failed to update monitor range")
		r.ErrorContains(err, t.Name())
	})
}

/*
func TestMonitor_Watch(t *testing.T) {
	r := require.New(t)
	p := &Persist{Path: dir(t)}
	r.NoError(p.Init())
	defer p.Close()

	c := &EthClient{
		Endpoint: "https://babel-api.testnet.iotex.io",
		Network:  NETWORK__IOTX_TESTNET,
	}
	r.NoError(c.Init())

	meta := Meta{
		Network:  NETWORK__IOTX_TESTNET,
		Contract: common.HexToAddress("0xCBb7a80983Fd3405972F700101A82DB6304C6547"),
		Topic:    common.HexToHash("0xa9ee0c223bc138bec6ebb21e09d00d5423fc3bbc210bdb6aef9d190b0641aecb"),
	}

	m := (&Monitor{Meta: meta}).
		WithPersistence(p).
		WithEthClient(c).
		WithStartBlock(25950000)
	r.NoError(m.Init())
	defer m.Stop()

	handler := func(sub Subscription, tx *types.Log) {}

	t.Run("WithoutOption", func(t *testing.T) {
		sub, err := m.Watch(nil, handler)
		r.NoError(err)
		go func() {
			err = <-sub.Err()
		}()
		sub.Unsubscribe()
		time.Sleep(time.Second)
		r.True(errors.Is(err, context.Canceled))
		r.False(slices.Contains(m.WatchList(), sub.ID()))
		_, err = p.Load(meta.WatcherEndKey(sub.ID()))
		r.True(errors.Is(err, pebble.ErrNotFound))
		_, err = p.Load(meta.WatcherFromKey(sub.ID()))
		r.True(errors.Is(err, pebble.ErrNotFound))
	})

	t.Run("WithOption", func(t *testing.T) {
		opt := &WatchOptions{SubID: "WithOption"}
		sub, err := m.Watch(opt, handler)
		r.NoError(err)
		r.Equal(sub.ID(), opt.SubID)

		sub2, err2 := m.Watch(opt, handler)
		r.Nil(sub2)
		r.ErrorContains(err2, "monitor is watching by subscriber")

		time.Sleep(time.Second)
		go func() {
			err = <-sub.Err()
		}()
		sub.Unsubscribe()
		time.Sleep(time.Second)
		r.True(errors.Is(err, context.Canceled))
		r.False(slices.Contains(m.WatchList(), sub.ID()))
		_, err = p.Load(meta.WatcherEndKey(sub.ID()))
		r.NoError(err)
		_, err = p.Load(meta.WatcherFromKey(sub.ID()))
		r.NoError(err)
	})

	t.Run("FailedToQueryWatcher", func(t *testing.T) {
		mock.Patch(p.QueryWatcher, func(Meta, string) (uint64, uint64, error) {
			return 0, 0, errors.New(t.Name())
		})
		sub, err := m.Watch(nil, handler)
		r.Nil(sub)
		r.ErrorContains(err, "failed to query watcher")
		r.ErrorContains(err, t.Name())
	})
	t.Run("FailedToUpdateWatcher", func(t *testing.T) {
		mock.Patch(p.UpdateWatcher, func(Meta, string, uint64, uint64) error {
			return errors.New(t.Name())
		})
		sub, err := m.Watch(nil, handler)
		r.Nil(sub)
		r.ErrorContains(err, "failed to update watcher")
		r.ErrorContains(err, t.Name())
	})
}

func ExampleMonitor() {
	path := filepath.Join(os.TempDir(), "ExampleMonitor")
	must.NoError(os.RemoveAll(path))
	must.NoError(os.MkdirAll(path, 0777))

	p := &Persist{Path: path}
	if err := p.Init(); err != nil {
		fmt.Println("persist init", err)
		return
	}
	defer p.Close()

	c := &EthClient{
		Endpoint: "https://babel-api.testnet.iotex.io",
		Network:  NETWORK__IOTX_TESTNET,
	}
	if err := c.Init(); err != nil {
		fmt.Println("client init", err)
		return
	}

	m := (&Monitor{
		Meta: Meta{
			Network:  NETWORK__IOTX_TESTNET,
			Contract: common.HexToAddress("0xCBb7a80983Fd3405972F700101A82DB6304C6547"),
			Topic:    common.HexToHash("0xa9ee0c223bc138bec6ebb21e09d00d5423fc3bbc210bdb6aef9d190b0641aecb"),
		},
	}).WithPersistence(p).
		WithEthClient(c).
		WithStartBlock(25952996)

	if err := m.Init(); err != nil {
		fmt.Println("monitor init", err)
		return
	}
	defer m.Stop()

	sig := make(chan struct{})
	_, err := m.Watch(
		&WatchOptions{
			SubID: "example",
			Start: ptrx.Ptr(uint64(25952996)),
		},
		func(sub Subscription, tx *types.Log) {
			if tx.BlockNumber >= 25954033 {
				fmt.Printf("consumed: block: %d hash: %s\n", tx.BlockNumber, tx.TxHash)
				sig <- struct{}{}
				return
			}
			fmt.Printf("consumed: block: %d hash: %s\n", tx.BlockNumber, tx.TxHash)
		})
	if err != nil {
		fmt.Println("sub", err)
		return
	}
	if len(m.WatchList()) != 1 {
		fmt.Println("check watch list failed")
		return
	}
	<-sig
	return

	// Output:
	// consumed: block: 25952996 hash: 0x29a3303a014ca27f5a0f9a76619958f4b60223e0ab59f78de70ea15810e0dbf6
	// consumed: block: 25953111 hash: 0x1ae26bdd08a1734a55c997bc54a87bc9dce8d0ce0229021063ec62bc353819ff
	// consumed: block: 25953227 hash: 0xcc42601dc3519b569d07c88273158b7a59e3aaf5b35e1984bf845805ef225262
	// consumed: block: 25953342 hash: 0xa0d8d7a9e16b4c33d2e12eba05383212e61bdcd4623da918b2b2097a022f57d8
	// consumed: block: 25953457 hash: 0x3ac0296b9eccc3e70d141818b7eac33b794c294599e441cedd4a47ce91e34580
	// consumed: block: 25953573 hash: 0xa41c6a454e3670c9980ef77724f4f30ac31fbf48ecfe1805b27734139fdc9c13
	// consumed: block: 25953688 hash: 0x05d5d5c70c2147ba9ac124a980631017532c71810609ef4714b0a23fddd3c932
	// consumed: block: 25953803 hash: 0x785a2c6e0b72e570e52646aa87821089a6b52c1a1defe99f42de9a16514af567
	// consumed: block: 25953918 hash: 0x21eb0a8adfd7006d2c8f04349d17adc56edc7ab2dbaf49863fa90a4c8edc5bc1
	// consumed: block: 25954033 hash: 0x5d1837c615ae094f8b5004d2166003feda59a2f52e4a2f3a10da4be222685e83
}
*/
