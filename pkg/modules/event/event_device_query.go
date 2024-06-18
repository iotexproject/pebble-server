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
	return (&TopicUnmarshaller{e, topic, "device", "query"}).Unmarshal()
}

func (e *DeviceQuery) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	dev := &models.Device{}
	err = FetchByPrimary(ctx, dev, e.imei)
	if err != nil {
		return err
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
		app := &models.App{}
		err = FetchByPrimary(ctx, app, parts[0])
		if err != nil {
			return err
		}
		firmware = app.ID
		uri = app.Uri
		version = app.Version
	}

	return PublicMqttMessage(ctx,
		"device_query", "backend/"+e.imei+"/status",
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
}
