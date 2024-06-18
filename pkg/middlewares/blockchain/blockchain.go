package blockchain

import (
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/xoctopus/x/misc/must"
)

type Blockchain struct {
	Clients     []*EthClient
	Contracts   []*Contract
	PersistPath string
	Network     Network
	AutoRun     bool

	monitors  sync.Map
	clients   map[Network]*EthClient
	contracts map[string]*Contract
	persist   TxPersistence
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

	if bc.AutoRun {
		err = bc.RunMonitors()
		return err
	}

	return nil
}

func (bc *Blockchain) Close() {
	bc.persist.Close()
}

func (bc *Blockchain) ClientByNetwork() *EthClient {
	return bc.clients[bc.Network]
}

func (bc *Blockchain) ContractByID(id string) *Contract {
	return bc.contracts[bc.Network.String()+"__"+id]
}

func (bc *Blockchain) Monitor(id, name string) *Monitor {
	contract := bc.ContractByID(id)
	if contract == nil {
		return nil
	}

	event, ok := contract.events[name]
	if !ok {
		return nil
	}

	meta := &Meta{contract.Network, contract.Address, event.event.ID}
	return must.BeTrueV(bc.monitors.Load(meta.MetaID())).(*Monitor)
}

func (bc *Blockchain) RunMonitors() error {
	for name, c := range bc.contracts {
		if c.Network != bc.Network {
			continue
		}
		for _, event := range c.events {
			monitor := &Monitor{
				Meta: Meta{
					Network:  c.Network,
					Contract: c.Address,
					Topic:    event.event.ID,
				},
				name:    name + "__" + event.Name,
				client:  bc.clients[c.Network],
				persist: bc.persist,
			}
			if _, ok := bc.monitors.Load(monitor.Meta); ok {
				continue
			}
			if err := monitor.Init(); err != nil {
				return errors.Wrapf(
					err, "failed to init monitor: [network:%s] [contract:%s] [topic:%s]",
					monitor.Network(), monitor.ContractAddress(), monitor.Topic(),
				)
			}
			bc.monitors.Store(monitor.meta, monitor)
		}
	}
	return nil
}

type MonitorInfo struct {
	Name        string         `json:"name"`
	Network     Network        `json:"network"`
	Endpoint    string         `json:"endpoint"`
	Contract    common.Address `json:"contract"`
	Topic       common.Hash    `json:"topic"`
	StartedAt   uint64         `json:"startedAt"`
	Current     uint64         `json:"current"`
	Subscribers []*struct {
		Name      string `json:"name"`
		StartedAt uint64 `json:"startedAt"`
		Current   uint64 `json:"current"`
	} `json:"subscribers"`
	meta MetaID
}

func (bc *Blockchain) MonitorsInfo() []*MonitorInfo {
	monitors := make([]*MonitorInfo, 0)
	bc.monitors.Range(func(_, v any) bool {
		m := v.(*Monitor)
		vv := &MonitorInfo{
			Name:     m.name,
			Network:  m.Network(),
			Endpoint: m.Endpoint(),
			Contract: m.ContractAddress(),
			Topic:    m.Topic(),
			meta:     m.MetaID(),
		}
		monitors = append(monitors, vv)
		m.subs.Range(func(key, _ any) bool {
			vv.Subscribers = append(vv.Subscribers, &struct {
				Name      string `json:"name"`
				StartedAt uint64 `json:"startedAt"`
				Current   uint64 `json:"current"`
			}{Name: key.(string), StartedAt: 0, Current: 0})
			return true
		})
		return true
	})

	for _, m := range monitors {
		start, current, _ := bc.persist.MetaRange(m.meta)
		m.StartedAt = start
		m.Current = current
		for _, s := range m.Subscribers {
			start, current, _ = bc.persist.QueryWatcher(m.meta, s.Name)
			s.StartedAt = start
			s.Current = current
		}
	}

	return monitors
}
