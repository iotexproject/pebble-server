package event

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
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

func (e *FirmwareUpdated) Unmarshal(v any) error {
	return v.(TxEventUnmarshaler).UnmarshalTx(e.EventName(), e)
}

func (e *FirmwareUpdated) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	app := &models.App{
		ID:             e.Name,
		Version:        e.Version,
		Uri:            e.Uri,
		Avatar:         e.Avatar,
		OperationTimes: models.NewOperationTimes(),
	}
	_, err = UpsertOnConflict(ctx, app, "id", "version", "uri", "avatar", "updated_at")
	if err != nil {
		return errors.Wrapf(err, "failed to upsert app: %s", app.ID)
	}

	meta, _ := contexts.AppMetaFromContext(ctx)
	err = PublicMqttMessage(ctx,
		"firmware_updated", "device/app_update/"+app.ID,
		&struct {
			Name       string `json:"name"`
			Version    string `json:"version"`
			Uri        string `json:"uri"`
			Avatar     string `json:"avatar"`
			ServerMeta string `json:"meta"`
		}{
			Name:       app.ID,
			Version:    app.Version,
			Uri:        app.Uri,
			Avatar:     app.Avatar,
			ServerMeta: meta.String(),
		},
	)
	return errors.Wrap(err, "failed to publish firmware_updated response")
}
