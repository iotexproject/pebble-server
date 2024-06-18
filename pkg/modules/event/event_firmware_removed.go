package event

import (
	"context"
	"strings"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
)

func init() {
	f := func() Event { return &FirmwareRemoved{} }
	e := f()
	registry(e.Topic(), f)
}

type FirmwareRemoved struct {
	Name string
}

func (e *FirmwareRemoved) Source() SourceType { return SOURCE_TYPE__BLOCKCHAIN }

func (e *FirmwareRemoved) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *FirmwareRemoved) ContractID() string { return enums.CONTRACT__PEBBLE_FIRMWARE }

func (e *FirmwareRemoved) EventName() string { return "FirmwareRemoved" }

func (e *FirmwareRemoved) Data() any { return e }

func (e *FirmwareRemoved) Unmarshal(any) error { return nil }

func (e *FirmwareRemoved) Handle(ctx context.Context) (err error) {
	return WrapHandleError(DeleteByPrimary(ctx, &models.App{}, e.Name), e)
}
