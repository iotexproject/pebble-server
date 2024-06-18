package event

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
)

func init() {
	f := func() Event { return &PebbleRemove{} }
	e := f()
	registry(e.Topic(), f)
}

type PebbleRemove struct {
	Imei  string
	Owner common.Address
}

func (e *PebbleRemove) Source() SourceType { return SOURCE_TYPE__BLOCKCHAIN }

func (e *PebbleRemove) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *PebbleRemove) ContractID() string { return enums.CONTRACT__PEBBLE_DEVICE }

func (e *PebbleRemove) EventName() string { return "Withdraw" }

func (e *PebbleRemove) Data() any { return e }

func (e *PebbleRemove) Unmarshal(any) error { return nil }

func (e *PebbleRemove) Handle(ctx context.Context) error {
	dev := &models.Device{}

	if dev.Owner != e.Owner.String() {
		return &HandleError{}
	}

	status := dev.Status
	if status == int32(models.CONFIRM) {
		status = int32(models.CREATED)
	}
	proposer := dev.Proposer
	if proposer == e.Owner.String() {
		proposer = ""
	}

	// update device set owner = '', status = $status, proposer = $proposer
	// where id = $e.imei

	return nil
}
