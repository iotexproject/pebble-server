package event

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
)

func init() {
	e := &DeviceQuery{}
	registry(e.Topic(), func() Event { return &DeviceQuery{} })
}

type DeviceQuery struct {
	IMEI
}

func (e *DeviceQuery) Source() SourceType { return SOURCE_TYPE__MQTT }

func (e *DeviceQuery) Topic() string { return "device/+/query" }

func (e *DeviceQuery) Unmarshal(any) error { return nil /* no payload */ }

func (e *DeviceQuery) UnmarshalTopic(topic []byte) error {
	return (&TopicParser{e, topic, "device", "query"}).Unmarshal()
}

func (e *DeviceQuery) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	dev := &models.Device{ID: e.imei}
	err = FetchByPrimary(ctx, dev)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch dev: %s", dev.ID)
	}

	if dev.Status == models.CREATED {
		return errors.Errorf("device %s is not propsaled", dev.ID)
	}

	var (
		firmware string
		uri      string
		version  string
	)
	if parts := strings.Split(dev.RealFirmware, " "); len(parts) == 2 {
		app := &models.App{ID: parts[0]}
		err = FetchByPrimary(ctx, app)
		if err != nil {
			return errors.Wrapf(err, "failed to fetch app: %s", app.ID)
		}
		firmware = app.ID
		uri = app.Uri
		version = app.Version
	}

	err = PublicMqttMessage(ctx,
		"device_query", "backend/"+e.imei+"/status", e.imei,
		&struct {
			Status   int32  `json:"status"`
			Proposer string `json:"proposer,omitempty"`
			Firmware string `json:"firmware,omitempty"`
			URI      string `json:"uri,omitempty"`
			Version  string `json:"version,omitempty"`
		}{
			Status:   dev.Status,
			Proposer: dev.Proposer,
			Firmware: firmware,
			URI:      uri,
			Version:  version,
		},
	)
	return errors.Wrapf(err, "failed to publish device_query response")
}
