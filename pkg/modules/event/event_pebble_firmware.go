package event

import (
	"context"
	"strings"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
)

func init() {
	f := func() Event {
		return &PebbleFirmware{
			contractID: enums.CONTRACT__PEBBLE_DEVICE,
		}
	}
	e := f()
	registry(e.Topic(), f)
}

type PebbleFirmware struct {
	Imei string
	App  string

	contractID string
}

func (e *PebbleFirmware) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *PebbleFirmware) Topic() string {
	return network.Topic(e.contractID) + "__" + strings.ToUpper(e.EventName())
}

func (e *PebbleFirmware) ContractID() string {
	return network.ContractID(e.contractID)
}

func (e *PebbleFirmware) EventName() string {
	return "Firmware"
}

func (e *PebbleFirmware) SubscriberID() string {
	return network.SubscriberID(e.contractID)
}

func (e *PebbleFirmware) Data() any { return e }

func (e *PebbleFirmware) Unmarshal(any) error { return nil }

func (e *PebbleFirmware) Handle(ctx context.Context) error {
	// app := select * from app where id = $appid
	// if app is not exist, return err
	// update device set firmware = '$app.id app.version' where id = $imei
	// notify device firmware updated
	// payload {firmware: $appid, uri: app.uri, version: app.version}
	// topic: backend/$imei/firmware
	return nil
}
