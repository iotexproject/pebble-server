package blockchain

import (
	"context"
	"math/big"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type MonitorClient interface {
	ChainID(context.Context) (*big.Int, error)
	ChainEndpoint() string
	BlockNumber(context.Context) (uint64, error)
	FilterLogs(context.Context, ethereum.FilterQuery) ([]types.Log, error)
}

func NewMonitor(network Network, contracts ...*Contract) *Monitor {
	meta := NewMeta(network, contracts...)
	return &Monitor{meta: *meta}
}

type Monitor struct {
	meta     Meta
	client   MonitorClient
	persist  TxPersistence
	subs     sync.Map
	interval time.Duration

	from    uint64
	current atomic.Uint64
	init    sync.Once
	started atomic.Bool

	stopped chan error
	cancel  context.CancelFunc
	stop    sync.Once
}

func (m *Monitor) Init() error {
	var err error
	m.init.Do(func() {
		var from, end uint64
		from, end, err = m.persist.MonitorRange(m.meta)
		if err != nil {
			err = errors.Wrap(err, "failed to load monitor range")
			return
		}

		if from == 0 && end == 0 {
			if m.from == 0 {
				m.from = 1
			}
			m.current.Store(m.from)
		} else {
			m.from = from
			m.current.Store(end)
		}
		if err = m.persist.UpdateMonitorRange(m.meta, m.from, m.current.Load()-1); err != nil {
			err = errors.Wrap(err, "failed to update monitor range")
			return
		}
		m.stopped = make(chan error)
		if m.interval == 0 {
			m.interval = 10 * time.Second
		}

		ctx, cancel := context.WithCancel(context.Background())
		m.cancel = cancel
		go m.run(ctx)
		m.started.Store(true)
		l.Info("monitor started", "from", m.from, "current", m.current.Load())
	})
	return err
}

func (m *Monitor) Stop() {
	m.stop.Do(func() {
		if m.started.Load() {
			m.cancel()
			err := <-m.stopped // wait monitor stop to safe close TxPersistence
			if errors.Is(err, context.Canceled) {
				l.Info("monitor stopped", "current", m.current.Load())
			} else {
				l.Error(err, "monitor failed", "current", m.current.Load())
			}
			m.subs.Range(func(k, v any) bool {
				v.(Subscription).Unsubscribe()
				return true
			})
		}
		m.started.Store(false)
	})
}

func (m *Monitor) WithPersistence(p TxPersistence) *Monitor {
	m.persist = p
	return m
}

func (m *Monitor) WithEthClient(c MonitorClient) *Monitor {
	m.client = c
	return m
}

func (m *Monitor) WithStartBlock(from uint64) *Monitor {
	m.from = from
	return m
}

func (m *Monitor) WithInterval(d time.Duration) *Monitor {
	m.interval = d
	return m
}

func (m *Monitor) StartAt() uint64 {
	return m.from
}

func (m *Monitor) CurrentBlock() uint64 {
	return m.current.Load()
}

func (m *Monitor) Endpoint() string {
	return m.client.ChainEndpoint()
}

func (m *Monitor) run(ctx context.Context) {
	filter := ethereum.FilterQuery{
		Addresses: m.meta.Contracts,
		Topics:    m.meta.Topics,
		FromBlock: new(big.Int),
		ToBlock:   new(big.Int),
	}
	step := uint64(1000)
	for {
		select {
		case <-ctx.Done():
			m.stopped <- ctx.Err()
			return
		default:
		}
		var (
			current uint64 // current block number
			from    uint64 // filter from
			to      uint64 // filter to
			logs    []*types.Log
			results []types.Log
			err     error
		)
		_, from, err = m.persist.MonitorRange(m.meta)
		if err != nil {
			goto Failed
		}
		from += 1 // current offset had be scanned
		current, err = m.client.BlockNumber(ctx)
		if err != nil {
			goto TryLater
		}
		current -= 1
		if from >= current {
			goto TryLater
		}
		to = min(current, from+step)
		filter.FromBlock.SetUint64(from)
		filter.ToBlock.SetUint64(to)
		results, err = m.client.FilterLogs(ctx, filter)
		if err != nil {
			goto TryLater
		}
		logs = make([]*types.Log, len(results))
		for i := range logs {
			logs[i] = &results[i]
		}
		if err = m.persist.InsertLogs(m.meta, logs...); err != nil {
			err = errors.Wrap(err, "failed to insert logs")
			goto Failed
		}
		if err = m.persist.UpdateMonitorRange(m.meta, m.from, to); err != nil {
			err = errors.Wrap(err, "failed to update meta range")
			goto Failed
		}
		l.Info("monitor queried", "from", from, "to", to, "count", len(logs))
		m.current.Store(to)
		if to == current {
			goto TryLater
		}
		continue
	TryLater:
		if err != nil && errors.Is(err, context.Canceled) {
			continue
		}
		time.Sleep(m.interval)
		continue
	Failed:
		m.stopped <- err
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

type EventHandler func(Subscription, *Contract, string, *types.Log)

func (m *Monitor) Watch(opts *WatchOptions, h EventHandler) (Subscription, error) {
	if opts == nil {
		opts = &WatchOptions{
			SubID:      uuid.NewString(),
			Start:      nil,
			unassigned: true,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := &subscriber{
		meta:    m.meta,
		id:      opts.SubID,
		err:     make(chan error, 1),
		stopped: make(chan error, 1),
		cancel:  cancel,
		cleanup: func() {
			m.subs.Delete(opts.SubID)
			if opts.unassigned {
				_ = m.persist.RemoveWatcher(m.meta, opts.SubID)
			}
		},
	}
	if _, loaded := m.subs.LoadOrStore(opts.SubID, s); loaded {
		return nil, errors.Errorf("subscriber `%s` watching", opts.SubID)
	}

	var (
		from uint64
		end  uint64
		err  error
	)

	defer func() {
		if err != nil {
			m.subs.Delete(opts.SubID)
		}
	}()

	from, end, err = m.persist.WatcherRange(m.meta, opts.SubID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query watcher")
	}
	if from == 0 && end == 0 {
		if opts.Start != nil {
			from = *opts.Start
		}
		if from == 0 {
			from = m.from
		}
		end = from
	}

	if err = m.persist.UpdateWatcherRange(m.meta, opts.SubID, from, end-1); err != nil {
		err = errors.Wrap(err, "failed to update watcher")
		return nil, err
	}

	w := &watcher{
		persist: m.persist,
		meta:    m.meta,
		sub:     s,
		name:    opts.SubID,
		from:    from,
		handler: h,
	}
	w.current.Store(end)
	go w.run(ctx)
	return s, nil
}
