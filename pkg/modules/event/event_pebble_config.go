package event

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
)

func init() {
	f := func() Event { return &PebbleConfig{} }
	e := f()
	registry(e.Topic(), f)
}

type PebbleConfig struct {
	IMEI
	Config string
}

func (e *PebbleConfig) Source() enums.EventSourceType {
	return enums.EVENT_SOURCE_TYPE__BLOCKCHAIN
}

func (e *PebbleConfig) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *PebbleConfig) ContractID() string { return enums.CONTRACT__PEBBLE_DEVICE }

func (e *PebbleConfig) EventName() string { return "Config" }

func (e *PebbleConfig) Unmarshal(v any) error {
	return v.(TxEventUnmarshaler).UnmarshalTx(e.EventName(), e)
}

func (e *PebbleConfig) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	if !contexts.IMEIFilter().MustFrom(ctx).NeedHandle(e.Imei) {
		return errors.Errorf("imei %s not in whitelist", e.Imei)
	}

	dev := &models.Device{ID: e.Imei, Config: e.Config}
	if err = UpdateByPrimary(ctx, dev, map[string]any{
		"config":     e.Config,
		"updated_at": time.Now(),
	}); err != nil {
		return errors.Wrapf(err, "failed to update device config: %s %s", dev.ID, dev.Config)
	}

	app := &models.AppV2{ID: e.Config}
	if err = FetchByPrimary(ctx, app); err != nil {
		return errors.Wrapf(err, "failed to fetch app_v2: %s", app.ID)
	}

	pubType := "pub_PebbleConfigRsp"
	pubData := app.Data
	return errors.Wrapf(
		PublicMqttMessage(ctx, pubType, "backend/"+e.Imei+"/config", pubData),
		"failed to publish %s", pubType,
	)
}
