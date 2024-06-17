package event

import (
	"context"
	"strings"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
)

func init() {
	f := func() Event {
		return &FirmwareRemoved{
			contractID: enums.CONTRACT__PEBBLE_FIRMWARE,
		}
	}
	e := f()
	registry(e.Topic(), f)
}

type FirmwareRemoved struct {
	Name string

	contractID string
}

func (e *FirmwareRemoved) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *FirmwareRemoved) Topic() string {
	return network.Topic(e.contractID) + "__" + strings.ToUpper(e.EventName())
}

func (e *FirmwareRemoved) ContractID() string {
	return network.ContractID(e.contractID)
}

func (e *FirmwareRemoved) EventName() string {
	return "FirmwareRemoved"
}

func (e *FirmwareRemoved) SubscriberID() string {
	return network.SubscriberID(e.contractID)
}

func (e *FirmwareRemoved) Data() any { return e }

func (e *FirmwareRemoved) Unmarshal(any) error { return nil }

func (e *FirmwareRemoved) Handle(ctx context.Context) error {
	// remove app by appid
	return nil
}
