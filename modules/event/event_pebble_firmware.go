package event

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
)

func init() {
	f := func() Event { return &PebbleFirmware{} }
	e := f()
	registry(e.Topic(), f)
}

type PebbleFirmware struct {
	IMEI
	App string
}

func (e *PebbleFirmware) Source() enums.EventSourceType {
	return enums.EVENT_SOURCE_TYPE__BLOCKCHAIN
}

func (e *PebbleFirmware) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *PebbleFirmware) ContractID() string { return enums.CONTRACT__PEBBLE_DEVICE }

func (e *PebbleFirmware) EventName() string { return "Firmware" }

func (e *PebbleFirmware) Unmarshal(v any) error {
	return v.(TxEventUnmarshaler).UnmarshalTx(e.EventName(), e)
}

func (e *PebbleFirmware) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	// if !contexts.IMEIFilter().MustFrom(ctx).NeedHandle(e.Imei) {
	// 	return errors.Errorf("imei %s not in whitelist", e.Imei)
	// }

	app := &models.App{ID: e.App}
	if err = FetchByPrimary(ctx, app); err != nil {
		return errors.Wrapf(err, "failed to fetch app: %s", app.ID)
	}

	dev := &models.Device{
		ID:       e.Imei,
		Firmware: app.ID + " " + app.Version,
	}
	err = UpdateByPrimary(ctx, dev, map[string]any{
		"firmware":   dev.Firmware,
		"updated_at": time.Now(),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to update device firmware: %s %s", dev.ID, dev.Firmware)
	}

	// meta := contexts.AppMeta().MustFrom(ctx)
	pubType := "pub_PebbleFirmwareRsp"
	pubData := &struct {
		Firmware string `json:"firmware"`
		Uri      string `json:"uri"`
		Version  string `json:"version"`
		// ServerMeta string `json:"server_meta"`
	}{
		Firmware: e.App,
		Uri:      app.Uri,
		Version:  app.Version,
		// ServerMeta: meta.String(),
	}

	return errors.Wrapf(
		PublicMqttMessage(ctx, pubType, "backend/"+e.Imei+"/firmware", pubData),
		"failed to publish %s", pubType,
	)
}
