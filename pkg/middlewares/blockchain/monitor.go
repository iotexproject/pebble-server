package blockchain

import (
	"bytes"
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/pkg/errors"
)

type MonitorMeta struct {
	Endpoint string
	Contract common.Address
	Topic    common.Hash
}

func (mi *MonitorMeta) ID() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(mi.Endpoint)
	buf.Write(mi.Contract.Bytes())
	buf.Write(mi.Topic.Bytes())

	h := crypto.Keccak256Hash(buf.Bytes())
	return h.String()
}

type Monitor struct {
	Metas []MonitorMeta

	metas map[string]*MonitorMeta
}

func (m *Monitor) Init() error {
	m.metas = make(map[string]*MonitorMeta)
	for _, meta := range m.Metas {
		id := meta.ID()
		if vm, exists := m.metas[id]; exists {
			return errors.Errorf("duplicated monitor meat: %s %s %s",
				vm.Endpoint, vm.Contract, vm.Topic)
		}
		m.metas[id] = &meta
	}
	return nil
}

type EventHandler func(instance *MonitorInstance, message *Message)

type MonitorInstance struct {
	ID string
	MonitorMeta
	From int64
	End  int64

	current int64
	client  *EthClient
	handler EventHandler
	persis  persistence
	stop    chan struct{}
}

func (m *MonitorInstance) Subscribe(handler EventHandler) error {
	m.handler = handler

	go func() {
		logs, err := m.persis.Query(m.ID, m.From, m.End)
		if err != nil {
			log.Error(err.Error())
			return
		}
		for _, l := range logs {
			m.handler(m, &Message{l})
			m.current = int64(l.BlockNumber)
		}
		if m.current < m.From {
			m.current = m.From
		}
		m.run()
	}()
	return nil
}

func (m *MonitorInstance) Unsubscribe() {
	m.stop <- struct{}{}
}

func (m *MonitorInstance) run() {
	query := ethereum.FilterQuery{
		Addresses: []common.Address{m.Contract},
		Topics:    [][]common.Hash{{m.Topic}},
	}
	interval := time.Second * 10
	step := int64(100000)
	for {
		select {
		case <-m.stop:
			log.Info("monitor stopped")
		default:
			latestBlk, err := m.client.BlockNumber(context.Background())
			if err != nil {
				log.Error("query latest block number", "msg", err)
				time.Sleep(interval)
				continue
			}
			log.Debug("query latest block", "block number", latestBlk)
			if uint64(m.current) > latestBlk {
				time.Sleep(interval)
				continue
			}
			query.FromBlock = big.NewInt(m.current)
			query.ToBlock = big.NewInt(min(m.current+step, int64(latestBlk)))

			logs, err := m.client.FilterLogs(context.Background(), query)
			if err != nil {
				log.Error("failed to filter logs", "msg", err)
				time.Sleep(interval)
				continue
			}
			log.Debug("filter logs", "from", query.FromBlock.Uint64(), "to", query.ToBlock.Uint64())
			m.current = query.ToBlock.Int64()
			if len(logs) == 0 {
				goto TryLater
			}
			log.Info("filter logs", "count", len(logs))
			for _, l := range logs {
				message := &Message{l}
				m.handler(m, message)
				m.persis.Upsert(m.ID, l)
			}
		TryLater:
			if query.ToBlock.Int64()-query.FromBlock.Int64() < step {
				time.Sleep(interval)
			}
		}
	}
}

type Message struct {
	types.Log
}

func (v *Message) Topic() string {
	return v.Log.Topics[0].String()
}
