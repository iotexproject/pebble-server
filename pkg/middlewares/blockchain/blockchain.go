package blockchain

import (
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

type Blockchain struct {
	Clients     []*EthClient
	Contracts   []*Contract
	Network     Network
	PersistPath string

	monitor   *Monitor
	clients   map[Network]*EthClient
	contracts map[string]*Contract
	persist   TxPersistence
	start     sync.Once
	stop      sync.Once
}

func (bc *Blockchain) SetDefault() {
	if len(bc.Clients) == 0 {
		bc.Clients = []*EthClient{
			{
				Network:  NETWORK__IOTX_TESTNET,
				Endpoint: "https://babel-api.testnet.iotex.io",
			},
			{
				Network:  NETWORK__IOTX_MAINNET,
				Endpoint: "https://babel-api.mainnet.iotex.io",
			},
		}
	}
	if bc.Network == NETWORK_UNKNOWN {
		bc.Network = NETWORK__IOTX_MAINNET
	}
	if bc.PersistPath == "" {
		bc.PersistPath = "/tmp/pebble"
	}
}

func (bc *Blockchain) Init() error {
	for _, c := range bc.Clients {
		if c.Network != bc.Network {
			continue
		}
		if err := c.Init(); err != nil {
			return err
		}
		if bc.clients == nil {
			bc.clients = make(map[Network]*EthClient)
		}
		for _, cc := range bc.clients {
			if cc.Network == c.Network || cc.Endpoint == c.Endpoint {
				return errors.Errorf(
					"client duplicated: [network:%s] [endpoint:%s]",
					c.Network, c.Endpoint,
				)
			}
		}
		bc.clients[c.Network] = c
	}

	sort.Slice(bc.Contracts, func(i, j int) bool {
		return bc.Contracts[i].ID < bc.Contracts[j].ID
	})

	for _, c := range bc.Contracts {
		if c.Network != bc.Network {
			continue
		}
		if err := c.Init(); err != nil {
			return errors.Wrapf(err, "failed to init contract: %s", c.ID)
		}
		if bc.contracts == nil {
			bc.contracts = make(map[string]*Contract)
		}
		for _, cc := range bc.contracts {
			if cc.ID == c.ID {
				return errors.Errorf("contract id `%s` duplicated", c.ID)
			}
			if cc.Network == c.Network && cc.Address == c.Address {
				return errors.Errorf(
					"contract `%s` duplicated with `%s` [network:%s] [address:%s]",
					c.ID, cc.ID, cc.Network, cc.Address,
				)
			}
		}
		bc.contracts[bc.Network.String()+"__"+c.ID] = c
	}

	var (
		err     error
		persist = &Persist{Path: bc.PersistPath}
	)
	if err = persist.Init(); err != nil {
		return errors.Wrapf(err, "failed to init bc persistence")
	}
	bc.persist = persist
	return nil
}

func (bc *Blockchain) Close() {
	bc.stop.Do(func() {
		if bc.monitor != nil {
			bc.monitor.Stop()
		}
		if bc.persist != nil {
			bc.persist.Close()
		}
	})
}

func (bc *Blockchain) ClientByNetwork() *EthClient {
	return bc.clients[bc.Network]
}

func (bc *Blockchain) ContractByID(id string) *Contract {
	return bc.contracts[bc.Network.String()+"__"+id]
}

func (bc *Blockchain) RunMonitor() error {
	var err error
	bc.start.Do(func() {
		bc.monitor = NewMonitor(bc.Network, bc.Contracts...)

		bc.monitor.WithInterval(10 * time.Second).
			WithStartBlock(14900000).
			WithEthClient(bc.clients[bc.Network]).
			WithPersistence(bc.persist)

		err = bc.monitor.Init()
	})
	return err
}

func (bc *Blockchain) Watch(options *WatchOptions, h EventHandler) (Subscription, error) {
	if err := bc.RunMonitor(); err != nil {
		return nil, err
	}
	return bc.monitor.Watch(options, h)
}

type MonitorInfo struct {
	Contracts map[string]struct {
		Address common.Address         `json:"address"`
		Events  map[string]common.Hash `json:"events"`
	} `json:"contracts"`
	From     uint64 `json:"from"`
	End      uint64 `json:"end"`
	Watchers []*struct {
		Sub  string `json:"sub"`
		From uint64 `json:"from"`
		End  uint64 `json:"end"`
	} `json:"watchers"`
}

func (bc *Blockchain) MonitorMeta() *MonitorInfo {
	info := &MonitorInfo{
		Contracts: make(map[string]struct {
			Address common.Address         `json:"address"`
			Events  map[string]common.Hash `json:"events"`
		}),
		From: bc.monitor.from,
		End:  bc.monitor.current.Load(),
	}
	for name, c := range bc.contracts {
		info.Contracts[name] = struct {
			Address common.Address         `json:"address"`
			Events  map[string]common.Hash `json:"events"`
		}{Address: c.Address, Events: make(map[string]common.Hash)}
		for _, v := range c.events {
			info.Contracts[name].Events[v.Name] = v.event.ID
		}
	}

	bc.monitor.subs.Range(func(key, _ any) bool {
		info.Watchers = append(info.Watchers, &struct {
			Sub  string `json:"sub"`
			From uint64 `json:"from"`
			End  uint64 `json:"end"`
		}{Sub: key.(string), From: 0, End: 0})
		return true
	})
	for _, s := range info.Watchers {
		s.From, s.End, _ = bc.persist.WatcherRange(bc.monitor.meta, s.Sub)
	}
	return info
}
