package blockchain

import (
	"context"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

type MonitorClient interface {
	ChainID(context.Context) (*big.Int, error)
	ChainEndpoint() string
	BlockNumber(context.Context) (uint64, error)
	FilterLogs(context.Context, ethereum.FilterQuery) ([]types.Log, error)
}

type Monitor struct {
	Meta    Meta
	meta    MetaID
	current atomic.Uint64
	client  MonitorClient
	persist TxPersistence
	cancel  context.CancelFunc
	subs    sync.Map
	name    string
}

func (m *Monitor) Init() error {
	m.meta = m.Meta.MetaID()

	_, end, err := m.persist.MetaRange(m.Meta.MetaID())
	if err != nil {
		return errors.Wrap(err, "failed to load monitor range")
	}

	m.current.Store(1)
	if end != 0 {
		m.current.Store(end)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go m.run(ctx)
	m.cancel = cancel

	l.Info("monitor started", m.fields()...)

	return nil
}

func (m *Monitor) fields(others ...any) []any {
	return append(others,
		"name", m.name,
		"current", m.current.Load(),
		"network", m.Network(),
		"endpoint", m.client.ChainEndpoint(),
		"contract", m.ContractAddress(),
		"topic", m.Topic(),
	)
}

func (m *Monitor) WithPersistence(p TxPersistence) *Monitor {
	m.persist = p
	return m
}

func (m *Monitor) WithEthClient(c MonitorClient) *Monitor {
	m.client = c
	return m
}

func (m *Monitor) MetaID() MetaID {
	if m.meta == [MetaIDLength]byte{} {
		m.meta = m.Meta.MetaID()
	}
	return m.meta
}

func (m *Monitor) CurrentBlock() uint64 {
	return m.current.Load()
}

func (m *Monitor) Network() Network {
	return m.Meta.Network
}

func (m *Monitor) Endpoint() string {
	return m.client.ChainEndpoint()
}

func (m *Monitor) ContractAddress() common.Address {
	return m.Meta.Contract
}

func (m *Monitor) Topic() common.Hash {
	return m.Meta.Topic
}

func (m *Monitor) Stop() {
	m.cancel()
}

func (m *Monitor) run(ctx context.Context) {
	filter := ethereum.FilterQuery{
		Addresses: []common.Address{m.Meta.Contract},
		Topics:    [][]common.Hash{{m.Meta.Topic}},
		FromBlock: new(big.Int),
		ToBlock:   new(big.Int),
	}
	interval := time.Second * 10
	step := uint64(100000)
	for {
		select {
		case <-ctx.Done():
			l.Info("monitor stopped", m.fields()...)
			return
		default:
		}
		var (
			highest uint64
			logs    []types.Log
			plogs   []*types.Log
			current uint64
			err     error
		)
		highest, err = m.client.BlockNumber(context.Background())
		if err != nil {
			goto TryLater
		}
		_, current, err = m.persist.MetaRange(m.meta)
		if err != nil {
			goto Failed
		}
		if current == 0 {
			current = 1
		}
		m.current.Store(current)
		if m.current.Load() > highest {
			goto TryLater
		}
		filter.FromBlock.SetUint64(m.current.Load())
		filter.ToBlock.SetUint64(min(m.current.Load()+step, highest))

		logs, err = m.client.FilterLogs(context.Background(), filter)
		if err != nil {
			goto TryLater
		}
		m.current.Store(filter.ToBlock.Uint64())
		if len(logs) > 0 {
			l.Info("monitor queried", m.fields("count", len(logs))...)
		}
		plogs = make([]*types.Log, len(logs))
		for i := range logs {
			plogs[i] = &logs[i]
		}
		if err = m.persist.InsertLogs(m.meta, plogs...); err != nil {
			err = errors.Wrap(err, "failed to insert logs")
			goto Failed
		}
		if err = m.persist.UpdateMetaRange(m.meta, 0, m.current.Load()); err != nil {
			err = errors.Wrap(err, "failed to update meta range")
			goto Failed
		}

		if new(big.Int).Sub(filter.ToBlock, filter.FromBlock).Uint64() == step {
			continue
		}
	TryLater:
		time.Sleep(interval)
		continue
	Failed:
		l.Error(err, "monitor failed", m.fields()...)
		return
	}
}

func (m *Monitor) WatchList() []string {
	subs := make([]string, 0)
	m.subs.Range(func(key, value any) bool {
		subs = append(subs, key.(string))
		return true
	})
	return subs
}

func (m *Monitor) Watch(opts WatchOptions, sink chan<- *types.Log) (Subscription, error) {
	if opts.SubID == "" {
		return nil, errors.Errorf("sub id is required")
	}
	if sink == nil {
		return nil, errors.Errorf("invalid sink channel, expect not nil")
	}

	if _, loaded := m.subs.LoadOrStore(opts.SubID, struct{}{}); loaded {
		return nil, errors.Errorf(
			"monitor is watching by subscriber `%s` [network: %s] [contract: %s] [topic:%s]",
			opts.SubID, m.Network(), m.ContractAddress(), m.Topic(),
		)
	}

	var (
		start   uint64
		current uint64
		err     error
	)

	defer func() {
		if err != nil {
			m.subs.Delete(opts.SubID)
		}
	}()

	start, current, err = m.persist.QueryWatcher(m.meta, opts.SubID)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			if opts.Start != nil {
				start = *opts.Start
				current = *opts.Start
				if start > 0 {
					current -= 1
				}
			}
		} else {
			err = errors.Wrap(err, "failed to query watcher")
			return nil, err
		}
	}

	if err = m.persist.UpdateWatcher(m.meta, opts.SubID, start, current); err != nil {
		err = errors.Wrap(err, "failed to update watcher")
		return nil, err
	}

	w := &watcher{
		persist: m.persist,
		meta:    m.Meta,
		metaID:  m.meta,
		sub:     opts.SubID,
		sink:    sink,
		start:   start,
	}
	return newSubscription(w, func() { m.subs.Delete(opts.SubID) }), nil
}
