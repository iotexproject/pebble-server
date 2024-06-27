package blockchain

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

type Subscription interface {
	ID() string
	Err() <-chan error
	Unsubscribe()
}

type WatchOptions struct {
	// SubID the unique id of subscriber, required field
	SubID string `json:"id"`
	// Start of the queried range(nil == latest) and this can be overwritten
	// by persisted record
	Start *uint64 `json:"start"`
	// unassigned means subscriber watch offset no need to persistent
	unassigned bool
}

type subscriber struct {
	meta    Meta
	id      string
	name    string
	err     chan error
	stopped chan error
	cancel  context.CancelFunc
	cleanup func()
}

func (s *subscriber) ID() string {
	return s.id
}

func (s *subscriber) Err() <-chan error {
	return s.err
}

func (s *subscriber) Unsubscribe() {
	s.cancel()
	err := <-s.stopped
	if errors.Is(err, context.Canceled) {
		l.Info("subscriber stopped", "name", s.name, "sub_id", s.id)
	} else {
		l.Error(err, "subscriber stopped", "name", s.name, "sub_id", s.id)
	}
	s.cleanup()
	s.err <- err
}

type watcher struct {
	persist TxPersistence
	meta    Meta
	sub     *subscriber
	name    string
	from    uint64
	current atomic.Uint64
	handler EventHandler
}

func (w *watcher) fields(others ...any) []any {
	return append(others,
		"current", w.current.Load(),
		"name", w.sub.name,
		"id", w.sub.id,
	)
}

func (w *watcher) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.sub.stopped <- ctx.Err()
			return
		default:
		}
		var (
			current uint64 // monitor current block height
			from    uint64 // scan from
			to      uint64 // scan to
			logs    []*types.Log
			err     error
		)
		_, from, err = w.persist.QueryWatcher(w.meta, w.name)
		if err != nil {
			goto Failed
		}
		from += 1
		_, current, err = w.persist.MetaRange(w.meta)
		if err != nil {
			goto Failed
		}
		if from >= current {
			goto TryLater
		}
		to = min(current, from+100000)
		logs, err = w.persist.QueryTxByHeightRange(w.meta, from, to)
		if err != nil {
			goto Failed
		}
		if w.handler != nil {
			for _, log := range logs {
				w.handler(w.sub, log)
			}
		}
		if err = w.persist.UpdateWatcher(w.meta, w.name, w.from, to); err != nil {
			goto Failed
		}
		if len(logs) > 0 {
			l.Info("watcher queried", w.fields("from", from, "to", to, "count", len(logs))...)
		}
		w.current.Store(to)
		if to == current {
			goto TryLater
		}
		continue
	TryLater:
		time.Sleep(time.Second * 10)
		continue
	Failed:
		w.sub.stopped <- err
		return
	}
}
