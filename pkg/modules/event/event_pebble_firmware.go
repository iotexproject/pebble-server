package event

import (
	"context"
	"strings"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
)

func init() {
	f := func() Event { return &PebbleFirmware{} }
	e := f()
	registry(e.Topic(), f)
}

type PebbleFirmware struct {
	Imei string
	App  string
}

func (e *PebbleFirmware) Source() SourceType { return SOURCE_TYPE__BLOCKCHAIN }

func (e *PebbleFirmware) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *PebbleFirmware) ContractID() string { return enums.CONTRACT__PEBBLE_DEVICE }

func (e *PebbleFirmware) EventName() string { return "Firmware" }

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
