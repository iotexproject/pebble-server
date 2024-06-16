package blockchain

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

type Contract struct {
	// ID human readable identifier in configuration
	ID string
	// Network deployed network
	Network Network
	// Address contract address
	Address common.Address
	// Events care about
	Events []*Event

	events map[string]*Event
}

func (c *Contract) Init() error {
	if c.ID == "" || c.Network == NETWORK_UNKNOWN {
		return errors.Errorf("invalid contract id or network")
	}
	for _, event := range c.Events {
		if err := event.Init(); err != nil {
			return errors.Wrapf(err, "failed to init event: `%s`", event.Name)
		}
		if c.events == nil {
			c.events = make(map[string]*Event)
		}
		_, ok := c.events[event.Name]
		if ok {
			return errors.Errorf("event `%s` duplicated", event.Name)
		}
		c.events[event.Name] = event
	}
	return nil
}

func (c *Contract) ParseTxLog(name string, log *types.Log, result any) (err error) {
	if c.Address != log.Address {
		return errors.Errorf("log address mismatched")
	}
	event, ok := c.events[name]
	if !ok {
		return errors.Errorf("event abi not found: %s", name)
	}
	if len(log.Topics) == 0 {
		return errors.Errorf("event signature not found in tx log")
	}
	if log.Topics[0] != event.event.ID {
		return errors.Errorf("event signature mismatched")
	}

	defer func() {
		v := recover()
		if v == nil {
			return
		}
		err = errors.Errorf("%v", v)
	}()

	if len(log.Data) > 0 {
		if err = event.target.UnpackIntoInterface(result, name, log.Data); err != nil {
			return errors.Wrapf(err, "failed to unpack")
		}
	}
	var indexed abi.Arguments
	for _, arg := range event.event.Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}

	err = abi.ParseTopics(result, indexed, log.Topics[1:])
	return err
}

func (c *Contract) Topic(name string) (common.Hash, bool) {
	ev, ok := c.events[name]
	if !ok {
		return common.Hash{}, false
	}
	return ev.event.ID, true
}

type Event struct {
	Name string
	ABI  string

	target abi.ABI   `env:"-"`
	event  abi.Event `env:"-"`
}

func (e *Event) Init() error {
	if len(e.ABI) == 0 || e.Name == "" {
		return errors.Errorf("invalid abi content or event name ")
	}

	if err := e.target.UnmarshalJSON([]byte(e.ABI)); err != nil {
		return errors.Wrapf(err, "failed to parse abi content to target")
	}
	event, ok := e.target.Events[e.Name]
	if !ok {
		return errors.Errorf("event `%s` not found in target abi", e.Name)
	}
	e.event = event
	return nil
}
