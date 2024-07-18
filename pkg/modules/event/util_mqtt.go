package event

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/xoctopus/x/misc/must"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
)

func PublicMqttMessage(ctx context.Context, tpe, topic string, v any) error {
	mq := must.BeTrueV(contexts.MqttBrokerFromContext(ctx))
	l := must.BeTrueV(contexts.LoggerFromContext(ctx))
	cli, err := mq.NewClient(uuid.NewString(), topic)
	if err != nil {
		return err
	}
	defer mq.Close(cli)

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
	if err = cli.Publish(data); err != nil {
		err = errors.Wrap(err, "failed to publish mqtt")
	}
	l.Info("mqtt published", "type", tpe, "topic", topic, "data", v, "result", err)
	return nil
}
