package event

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/xoctopus/x/misc/must"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
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
	mq := must.BeTrueV(contexts.MqttBrokerFromContext(ctx))

	defer func() { err = WrapHandleError(err, e) }()

	dev, err := FetchDeviceByIMEI(ctx, e.imei)
	if err != nil {
		return err
	}

	if dev.Status == int32(models.CREATED) {
		return errors.Errorf("device %s is not propsaled", dev.ID)
	}

	var (
		firmware string
		uri      string
		version  string
	)
	if parts := strings.Split(dev.RealFirmware, " "); len(parts) == 2 {
		app, err := FetchFirmwareByID(ctx, parts[0])
		if err != nil {
			return err
		}
		firmware = app.ID
		uri = app.Uri
		version = app.Version
	}

	cli, err := mq.NewClient(
		"device_query_rsp",
		strings.Join([]string{"backend", e.imei, "status"}, "/"),
	)
	if err != nil {
		return err
	}
	defer mq.Close(cli)

	err = cli.Publish(must.NoErrorV(json.Marshal(&struct {
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
	})))
	return
}
