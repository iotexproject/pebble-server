package event

import (
	"context"
	"strings"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/ethutil/address"
)

func init() {
	f := func() Event {
		return &PebbleRemove{
			contractID: enums.CONTRACT__PEBBLE_DEVICE,
		}
	}
	e := f()
	registry(e.Topic(), f)
}

type PebbleRemove struct {
	imei  string
	owner address.Address

	contractID string
}

func (e *PebbleRemove) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *PebbleRemove) Topic() string {
	return network.Topic(e.contractID) + "__" + strings.ToUpper(e.EventName())
}

func (e *PebbleRemove) ContractID() string {
	return network.ContractID(e.contractID)
}

func (e *PebbleRemove) EventName() string {
	return "Withdraw"
}

func (e *PebbleRemove) SubscriberID() string {
	return network.SubscriberID(e.contractID)
}

func (e *PebbleRemove) Data() any { return e }

func (e *PebbleRemove) Unmarshal(any) error { return nil }

func (e *PebbleRemove) Handle(ctx context.Context) error {
	dev := &models.Device{}

	if dev.Owner != e.owner.String() {
		return &HandleError{}
	}

	status := dev.Status
	if status == int32(models.CONFIRM) {
		status = int32(models.CREATED)
	}
	proposer := dev.Proposer
	if proposer == e.owner.String() {
		proposer = ""
	}

	// update device set owner = '', status = $status, proposer = $proposer
	// where id = $e.imei

	return nil
}
