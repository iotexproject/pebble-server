package event

import (
	"context"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/ethutil/address"
)

func init() {
	e := &PebbleConfirmed{}
	registry(e.Topic(), func() Event { return &PebbleConfirmed{} })
}

type PebbleConfirmed struct {
	imei    string
	owner   address.Address
	device  address.Address
	channel uint32
}

func (e *PebbleConfirmed) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *PebbleConfirmed) Topic() string {
	return "Confirm(string imei, address owner, address device, uint32 channel)"
}

func (e *PebbleConfirmed) Unmarshal(data []byte) error {
	// unmarshal event log
	return nil
}

func (e *PebbleConfirmed) Handle(ctx context.Context) error {
	// insert into device
	// id=$imei,owner=$owner,address=$device,proposer='',status=CONFIRM
	// created_at,updated_at
	// on conflict update owner,proposer,status,updated_at
	return nil
}
