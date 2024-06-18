package blockchain

import (
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

type Subscription interface {
	Err() <-chan error
	Unsubscribe()
}

type WatchOptions struct {
	// SubID the unique id of subscriber, required field
	SubID string `json:"id"`
	// Start of the queried range(nil == latest) and this can be overwritten
	// by persisted record
	Start *uint64 `json:"start"`
}

func newSubscription(w *watcher, cleanup func()) Subscription {
	err := make(chan error, 1)
	stop := make(chan struct{}, 1)
	s := &subscriber{
		stop:    stop,
		cleanup: cleanup,
		err:     err,
	}
	w.stop = stop
	w.err = err

	l.Info("watcher started", w.fields()...)
	go w.run()
	return s
}

type subscriber struct {
	stop    chan<- struct{}
	err     chan error
	cleanup func()
}

func (s *subscriber) Err() <-chan error {
	return s.err
}

func (s *subscriber) Unsubscribe() {
	s.stop <- struct{}{}
	s.cleanup()
	s.err <- errors.Errorf("user unsubscribed")
}

type watcher struct {
	persist TxPersistence
	meta    Meta
	metaID  MetaID
	sub     string
	stop    <-chan struct{}
	sink    chan<- *types.Log
	err     chan<- error
	start   uint64
	current uint64
}

func (w *watcher) fields(others ...any) []any {
	return append(others,
		"start", w.start,
		"current", w.current,
		"subscriber", w.sub,
		"network", w.meta.Network,
		"contract", w.meta.Contract,
		"topic", w.meta.Topic,
	)
}

func (w *watcher) run() {
	step := uint64(100000)
	for {
		select {
		case <-w.stop:
			l.Info("watcher stopped", w.fields()...)
			return
		default:
			var (
				logs    []*types.Log
				highest uint64
				next    uint64
				err     error
			)
			_, w.current, err = w.persist.QueryWatcher(w.metaID, w.sub)
			if err != nil {
				goto Failed
			}
			w.current += 1
			_, highest, err = w.persist.MetaRange(w.metaID)
			if err != nil {
				goto Failed
			}
			if w.current > highest {
				goto TryLater // wait monitor sync
			}
			next = min(highest, w.current+step)
			logs, err = w.persist.QueryTxByHeightRange(w.metaID, w.current, next)
			if err != nil {
				goto Failed
			}
			w.current = next
			if len(logs) > 0 {
				l.Info("watcher queried", w.fields("count", len(logs))...)
			}
			for _, log := range logs {
				w.sink <- log
			}
			err = w.persist.UpdateWatcher(w.metaID, w.sub, w.start, next)
			if err != nil {
				goto Failed
			}
			if next < highest {
				continue
			}
		TryLater:
			time.Sleep(time.Second * 10)
			continue
		Failed:
			w.err <- err
			l.Error(err, "watcher stopped", w.fields()...)
			return
		}
	}
}
