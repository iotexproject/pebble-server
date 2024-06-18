package event

import (
	"context"
	"strings"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
)

func init() {
	f := func() Event { return &FirmwareUpdated{} }
	e := f()
	registry(e.Topic(), f)
}

type FirmwareUpdated struct {
	Name    string
	Version string
	Uri     string
	Avatar  string
}

func (e *FirmwareUpdated) Source() SourceType { return SOURCE_TYPE__BLOCKCHAIN }

func (e *FirmwareUpdated) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *FirmwareUpdated) ContractID() string { return enums.CONTRACT__PEBBLE_FIRMWARE }

func (e *FirmwareUpdated) EventName() string { return "FirmwareUpdated" }

func (e *FirmwareUpdated) Data() any { return e }

func (e *FirmwareUpdated) Unmarshal(any) error { return nil }

func (e *FirmwareUpdated) Handle(ctx context.Context) error {
	// create or update app
	// notify device firmware updated device/app_updated/$appid
	// {name:app.id,version:app.version,uri:app.uri,avatar:app.avatar}
	return nil
}
