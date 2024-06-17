package event

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
)

func init() {
	f := func() Event {
		return &PebbleProposal{
			contractID: enums.CONTRACT__PEBBLE_DEVICE,
		}
	}
	e := f()
	registry(e.Topic(), f)
}

type PebbleProposal struct {
	Imei   string
	Owner  common.Address
	Device common.Address
	Name   string
	Avatar string

	contractID string
}

func (e *PebbleProposal) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *PebbleProposal) Topic() string {
	return network.Topic(e.contractID) + "__" + strings.ToUpper(e.EventName())
}

func (e *PebbleProposal) ContractID() string {
	return network.ContractID(e.contractID)
}

func (e *PebbleProposal) EventName() string {
	return "Proposal"
}

func (e *PebbleProposal) SubscriberID() string {
	return network.SubscriberID(e.contractID)
}

func (e *PebbleProposal) Data() any { return e }

func (e *PebbleProposal) Unmarshal(any) error { return nil }

func (e *PebbleProposal) Handle(ctx context.Context) error {
	// insert into device
	// id=$imei,address=$device,proposer='',name=$name,avatar=$avatar,status=PROPOSAL
	// created_at,updated_at
	// on conflict update proposer,name,avatar,status,updated_at
	return nil
}
