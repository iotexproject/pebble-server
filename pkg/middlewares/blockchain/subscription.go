package blockchain

import (
	"context"
	"sync/atomic"
	"time"

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
		l.Info("watcher stopped", "sub_id", s.id)
	} else {
		l.Error(err, "watcher stopped", "sub_id", s.id)
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
		"sub_id", w.sub.id,
		"current", w.current.Load(),
	)
}

func (w *watcher) run(ctx context.Context) {
	step := uint64(1000)
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
			results []*result
			err     error
		)
		_, from, err = w.persist.WatcherRange(w.meta, w.name)
		if err != nil {
			goto Failed
		}
		from += 1
		_, current, err = w.persist.MonitorRange(w.meta)
		if err != nil {
			goto Failed
		}
		if from >= current {
			goto TryLater
		}
		to = min(current, from+step)
		results, err = w.persist.QueryTxByHeightRange(w.meta, from, to)
		if err != nil {
			goto Failed
		}
		if w.handler != nil {
			for _, res := range results {
				key := string(res.key)
				c, event, err := w.meta.ParseLogKey(key)
				if err != nil {
					l.Error(err, "failed to parse log key: %s", key)
					continue
				}
				w.handler(w.sub, c, event, res.log)
			}
		}
		if err = w.persist.UpdateWatcherRange(w.meta, w.name, w.from, to); err != nil {
			goto Failed
		}
		l.Info("watcher queried", w.fields("from", from, "to", to, "count", len(results))...)
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
