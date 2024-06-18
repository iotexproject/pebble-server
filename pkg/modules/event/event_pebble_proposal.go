package event

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
)

func init() {
	f := func() Event { return &PebbleProposal{} }
	e := f()
	registry(e.Topic(), f)
}

type PebbleProposal struct {
	Imei   string
	Owner  common.Address
	Device common.Address
	Name   string
	Avatar string
}

func (e *PebbleProposal) Source() SourceType { return SOURCE_TYPE__BLOCKCHAIN }

func (e *PebbleProposal) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *PebbleProposal) ContractID() string { return enums.CONTRACT__PEBBLE_DEVICE }

func (e *PebbleProposal) EventName() string { return "Proposal" }

func (e *PebbleProposal) Data() any { return e }

func (e *PebbleProposal) Unmarshal(any) error { return nil }

func (e *PebbleProposal) Handle(ctx context.Context) error {
	// insert into device
	// id=$imei,address=$device,proposer='',name=$name,avatar=$avatar,status=PROPOSAL
	// created_at,updated_at
	// on conflict update proposer,name,avatar,status,updated_at
	return nil
}
