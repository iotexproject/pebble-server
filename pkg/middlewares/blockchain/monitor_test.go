package blockchain_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xoctopus/x/misc/must"
	"github.com/xoctopus/x/ptrx"

	. "github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/blockchain"
)

func TestMonitor_Init(t *testing.T) {
	r := require.New(t)

	p := &Persist{Path: dir(t)}
	r.NoError(p.Init())

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
		WithEthClient(c)
	// force monitor start at 25950000
	r.NoError(p.SetUint64(MetaRangeEndKey(m.MetaID()), 25950000))

	r.Equal(m.MetaID(), meta.MetaID())
	r.Equal(m.Network(), meta.Network)
	r.Equal(m.Endpoint(), c.ChainEndpoint())
	r.Equal(m.Topic().String(), "0xa9ee0c223bc138bec6ebb21e09d00d5423fc3bbc210bdb6aef9d190b0641aecb")
	r.Equal(m.ContractAddress().String(), "0xCBb7a80983Fd3405972F700101A82DB6304C6547")

	t.Run("Success", func(t *testing.T) {
		r.NoError(m.Init())
		m.Stop()
	})
	t.Run("FailedToLoadMonitorRange", func(t *testing.T) {
		mock.Patch(p.DB().Get, func([]byte) ([]byte, io.Closer, error) {
			return nil, nil, errors.New(t.Name())
		})
		err := m.Init()
		r.ErrorContains(err, "failed to load monitor range")
		r.ErrorContains(err, t.Name())
	})
	t.Run("OverwriteCurrentFromPersistence", func(t *testing.T) {
		ori := m.CurrentBlock()
		r.NoError(p.SetUint64(MetaRangeEndKey(m.MetaID()), ori-100))
		r.NoError(m.Init())
		r.Equal(m.CurrentBlock(), ori-100)
		m.Stop()
	})
}

func TestMonitor_Watch(t *testing.T) {
	r := require.New(t)
	p := &Persist{Path: dir(t)}
	r.NoError(p.Init())

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
		WithEthClient(c)
	// force monitor start at 25950000
	r.NoError(p.SetUint64(MetaRangeEndKey(m.MetaID()), 25950000))

	r.NoError(m.Init())

	t.Run("InvalidSubID", func(t *testing.T) {
		opts := WatchOptions{}
		sub, err := m.Watch(opts, make(chan *types.Log, 10))
		r.Nil(sub)
		r.ErrorContains(err, "sub id is required")
	})
	t.Run("InvalidSink", func(t *testing.T) {
		opts := WatchOptions{SubID: "any"}
		sub, err := m.Watch(opts, nil)
		r.Nil(sub)
		r.ErrorContains(err, "invalid sink channel")
	})
	t.Run("FailedToQueryWatcher", func(t *testing.T) {
		mock.Patch(p.Load, func([]byte) ([]byte, error) {
			return nil, errors.New(t.Name())
		})
		opts := WatchOptions{SubID: "any", Start: ptrx.Ptr(uint64(100))}
		sub, err := m.Watch(opts, make(chan *types.Log, 10))
		r.Nil(sub)
		r.ErrorContains(err, "failed to query watcher")
		r.ErrorContains(err, t.Name())
		r.Len(m.WatchList(), 0)
	})
	t.Run("FailedToUpdateWatcher", func(t *testing.T) {
		mock.Patch(p.Store, func([]byte, []byte) error {
			return errors.New(t.Name())
		})
		opts := WatchOptions{SubID: "any", Start: ptrx.Ptr(uint64(100))}
		sub, err := m.Watch(opts, make(chan *types.Log, 10))
		r.Nil(sub)
		r.ErrorContains(err, "failed to update watcher")
		r.ErrorContains(err, t.Name())
		r.Len(m.WatchList(), 0)
	})
	t.Run("Subscribed", func(t *testing.T) {
		consume := func(name string, sub Subscription, sink <-chan *types.Log) {
			defer sub.Unsubscribe()
			for {
				select {
				case err := <-sub.Err():
					t.Logf("%s: %v", name, err)
					return
				default:
					l := <-sink
					t.Logf("consumed: blk: %d hash: %s", l.BlockNumber, l.TxHash)
					return
				}
			}
		}

		opts := WatchOptions{SubID: "sub1", Start: ptrx.Ptr(uint64(100))}
		sink1 := make(chan *types.Log, 100)
		sub1, err := m.Watch(opts, sink1)
		r.NoError(err)
		r.Equal(m.WatchList(), []string{"sub1"})
		go consume("sub1", sub1, sink1)

		sink2 := make(chan *types.Log, 100)
		sub2, err := m.Watch(opts, sink2)
		r.Nil(sub2)
		r.ErrorContains(err, "monitor is watching by subscriber `sub1`")
		sub1.Unsubscribe()
		r.Equal(m.WatchList(), []string{})
	})

	time.Sleep(time.Second)
	m.Stop()
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
	}).WithPersistence(p).WithEthClient(c)
	// force monitor start at 25952995(on 25952996 block height contains tx monitor cares)
	if err := p.SetUint64(MetaRangeEndKey(m.MetaID()), 25952995); err != nil {
		fmt.Println("force set monitor start", err)
		return
	}

	if err := m.Init(); err != nil {
		fmt.Println("monitor init", err)
		return
	}

	consume := func(name string, sub Subscription, sink <-chan *types.Log) {
		defer sub.Unsubscribe()
		for {
			select {
			case err := <-sub.Err():
				fmt.Printf("%s: %v\n", name, err)
				return
			default:
				l := <-sink
				fmt.Printf("consumed: block: %d hash: %s\n", l.BlockNumber, l.TxHash)
				if l.BlockNumber == 25954033 {
					return
				}
			}
		}
	}

	sink := make(chan *types.Log, 10)
	sub, err := m.Watch(WatchOptions{SubID: "example", Start: ptrx.Ptr(uint64(25952000))}, sink)
	if err != nil {
		fmt.Println("sub", err)
		return
	}
	consume("example", sub, sink)
	if len(m.WatchList()) != 0 {
		fmt.Println("check watch list failed")
		return
	}

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
