package blockchain

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/xoctopus/x/misc/must"
)

type Blockchain struct {
	Clients     []*EthClient
	Contracts   []*Contract
	PersistPath string

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
	if bc.PersistPath == "" {
		bc.PersistPath = "/tmp/pebble"
	}
}

func (bc *Blockchain) Init() error {
	for _, c := range bc.Clients {
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
		if bc.clients[c.Network] == nil {
			return errors.Errorf("contract network `%d` not found", c.Network)
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
		bc.contracts[c.ID] = c
	}

	persist := &Persist{Path: bc.PersistPath}
	if err := persist.Init(); err != nil {
		return errors.Wrapf(err, "failed to init persistence")
	}
	bc.persist = persist

	for _, c := range bc.contracts {
		for _, event := range c.events {
			monitor := &Monitor{
				Meta: Meta{
					Network:  c.Network,
					Contract: c.Address,
					Topic:    event.event.ID,
				},
				client:  bc.clients[c.Network],
				persist: bc.persist,
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

func (bc *Blockchain) ClientByNetwork(network Network) *EthClient {
	return bc.clients[network]
}

func (bc *Blockchain) ContractByID(id string) *Contract {
	return bc.contracts[id]
}

func (bc *Blockchain) Monitor(contractID, eventName string) *Monitor {
	contract := bc.contracts[contractID]
	if contract == nil {
		return nil
	}

	event, ok := contract.events[eventName]
	if !ok {
		return nil
	}

	meta := &Meta{contract.Network, contract.Address, event.event.ID}
	return must.BeTrueV(bc.monitors.Load(meta.MetaID())).(*Monitor)
}
