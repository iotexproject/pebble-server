package event

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
)

func init() {
	f := func() Event { return &PebbleConfig{} }
	e := f()
	registry(e.Topic(), f)
}

type PebbleConfig struct {
	Imei   string
	Config string
}

func (e *PebbleConfig) Source() SourceType { return SOURCE_TYPE__BLOCKCHAIN }

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

	md := &models.Device{ID: e.Imei}
	fs := map[string]any{"config": e.Config}
	if err = UpdateByPrimary(ctx, md, fs); err != nil {
		return err
	}

	app := &models.AppV2{ID: e.Config}
	if err = FetchByPrimary(ctx, app); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	return PublicMqttMessage(ctx,
		"pebble_config",
		"backend/"+e.Imei+"/config",
		app.Data,
	)
}
