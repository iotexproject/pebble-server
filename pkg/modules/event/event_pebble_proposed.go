package event

import (
	"context"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/ethutil/address"
)

func init() {
	e := &PebbleProposed{}
	registry(e.Topic(), func() Event { return &PebbleProposed{} })
}

type PebbleProposed struct {
	imei   string
	owner  address.Address
	device address.Address
	name   string
	avatar string
}

func (e *PebbleProposed) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *PebbleProposed) Topic() string {
	return "Proposal(string imei, address owner, address device, string name, string avatar)"
}

func (e *PebbleProposed) Unmarshal(v any) error {
	// unmarshal event log
	return nil
}

func (e *PebbleProposed) Handle(ctx context.Context) error {
	// insert into device
	// id=$imei,address=$device,proposer='',name=$name,avatar=$avatar,status=PROPOSAL
	// created_at,updated_at
	// on conflict update proposer,name,avatar,status,updated_at
	return nil
}
