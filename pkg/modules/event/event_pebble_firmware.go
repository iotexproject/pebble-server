package event

import (
	"context"
	"strings"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
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

func (e *PebbleFirmware) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	app := &models.App{ID: e.App}
	if err = FetchByPrimary(ctx, app, e.Imei); err != nil {
		return err
	}

	dev := &models.Device{
		ID:       e.Imei,
		Firmware: app.ID + " " + app.Version,
	}
	err = UpdateByPrimary(ctx, dev, e.Imei, map[string]any{"firmware": dev.Firmware})
	if err != nil {
		return err
	}

	return PublicMqttMessage(ctx,
		"pebble_firmware", "backend/"+e.Imei+"/firmware",
		&struct {
			Firmware string `json:"firmware"`
			Uri      string `json:"uri"`
			Version  string `json:"version"`
		}{
			Firmware: e.App,
			Uri:      app.Uri,
			Version:  app.Version,
		},
	)
	return nil
}
