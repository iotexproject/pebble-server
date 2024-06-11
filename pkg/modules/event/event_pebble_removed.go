package event

import (
	"context"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/ethutil/address"
)

func init() {
	e := &PebbleRemoved{}
	registry(e.Topic(), func() Event { return &PebbleRemoved{} })
}

type PebbleRemoved struct {
	imei  string
	owner address.Address
}

func (e *PebbleRemoved) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *PebbleRemoved) Topic() string {
	return "Remove(string imei, address owner)"
}

func (e *PebbleRemoved) Unmarshal(data []byte) error {
	// unmarshal event log
	return nil
}

func (e *PebbleRemoved) Handle(ctx context.Context) error {
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
