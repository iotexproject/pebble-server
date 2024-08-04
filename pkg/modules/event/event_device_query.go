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
	e := &DeviceQuery{}
	registry(e.Topic(), func() Event { return &DeviceQuery{} })
}

type DeviceQuery struct {
	IMEI
}

func (e *DeviceQuery) Source() enums.EventSourceType {
	return enums.EVENT_SOURCE_TYPE__MQTT
}

func (e *DeviceQuery) Topic() string { return "device/+/query" }

func (e *DeviceQuery) Unmarshal(any) error { return nil /* no payload */ }

func (e *DeviceQuery) UnmarshalTopic(topic []byte) error {
	return (&TopicParser{e, topic, "device", "query"}).Unmarshal()
}

func (e *DeviceQuery) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	if !contexts.IMEIFilter().MustFrom(ctx).NeedHandle(e.Imei) {
		return errors.Errorf("imei %s not in whitelist", e.Imei)
	}

	dev := &models.Device{ID: e.Imei}
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

	// meta := contexts.AppMeta().MustFrom(ctx)
	pubType := "pub_DeviceQueryRsp"
	pubData := &struct {
		Status     int32  `json:"status"`
		Proposer   string `json:"proposer,omitempty"`
		Firmware   string `json:"firmware,omitempty"`
		URI        string `json:"uri,omitempty"`
		Version    string `json:"version,omitempty"`
		ServerMeta string `json:"server_meta,omitempty"`
	}{
		Status:   dev.Status,
		Proposer: dev.Proposer,
		Firmware: firmware,
		URI:      uri,
		Version:  version,
		// ServerMeta: meta.String(),
	}
	return errors.Wrapf(
		PublicMqttMessage(ctx, pubType, "backend/"+e.Imei+"/status", pubData),
		"failed to publish %s", pubType,
	)
}
