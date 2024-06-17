package event

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
)

func init() {
	f := func() Event {
		return &PebbleConfirm{
			contractID: enums.CONTRACT__PEBBLE_DEVICE,
		}
	}
	e := f()
	registry(e.Topic(), f)
}

type PebbleConfirm struct {
	Imei    string
	Owner   common.Address
	Device  common.Address
	Channel uint32

	contractID string
}

func (e *PebbleConfirm) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *PebbleConfirm) Topic() string {
	return network.Topic(e.contractID) + "__" + strings.ToUpper(e.EventName())
}

func (e *PebbleConfirm) ContractID() string {
	return network.ContractID(e.contractID)
}

func (e *PebbleConfirm) EventName() string {
	return "Confirm"
}

func (e *PebbleConfirm) SubscriberID() string {
	return network.SubscriberID(e.contractID)
}

func (e *PebbleConfirm) Data() any { return e }

func (e *PebbleConfirm) Unmarshal(any) error { return nil }

func (e *PebbleConfirm) Handle(ctx context.Context) error {
	// insert into device
	// id=$imei,owner=$owner,address=$device,proposer='',status=CONFIRM
	// created_at,updated_at
	// on conflict update owner,proposer,status,updated_at
	return nil
}
