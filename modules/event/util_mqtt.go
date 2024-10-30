package event

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
)

func PublicMqttMessage(ctx context.Context, tpe, topic string, v any) error {
	clientID, _ := contexts.MqttClientID().From(ctx)
	if clientID == "" {
		clientID = uuid.NewString()
	}
	clientID = "pub_" + clientID

	mq := contexts.MqttBroker().MustFrom(ctx)
	l := contexts.Logger().MustFrom(ctx)
	cli, err := mq.NewClient(clientID, topic)
	if err != nil {
		return err
	}

	var data []byte
	switch _v := v.(type) {
	case string:
		data = []byte(_v)
	case []byte:
		data = _v
	default:
		data, err = json.Marshal(_v)
		if err != nil {
			return err
		}
	}

	if ok, _ := contexts.DryRun().From(ctx); ok {
		return nil
	}
	if err = cli.Publish(data); err != nil {
		err = errors.Wrap(err, "failed to publish mqtt")
	}
	l.Info("mqtt published", "type", tpe, "client", clientID, "topic", topic, "data", v, "result", err)
	return nil
}
