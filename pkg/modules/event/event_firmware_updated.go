package event

import (
	"context"
	"strings"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
)

func init() {
	f := func() Event {
		return &FirmwareUpdated{
			contractID: enums.CONTRACT__PEBBLE_FIRMWARE,
		}
	}
	e := f()
	registry(e.Topic(), f)
}

type FirmwareUpdated struct {
	Name    string
	Version string
	Uri     string
	Avatar  string

	contractID string
}

func (e *FirmwareUpdated) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *FirmwareUpdated) Topic() string {
	return network.Topic(e.contractID) + "__" + strings.ToUpper(e.EventName())
}

func (e *FirmwareUpdated) ContractID() string {
	return network.ContractID(e.contractID)
}

func (e *FirmwareUpdated) EventName() string {
	return "FirmwareUpdated"
}

func (e *FirmwareUpdated) SubscriberID() string {
	return network.SubscriberID(e.contractID)
}

func (e *FirmwareUpdated) Data() any { return e }

func (e *FirmwareUpdated) Unmarshal(any) error { return nil }

func (e *FirmwareUpdated) Handle(ctx context.Context) error {
	// create or update app
	// notify device firmware updated device/app_updated/$appid
	// {name:app.id,version:app.version,uri:app.uri,avatar:app.avatar}
	return nil
}
