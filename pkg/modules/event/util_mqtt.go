package event

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/xoctopus/x/misc/must"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
)

func PublicMqttMessage(ctx context.Context, id, topic string, v any) error {
	mq := must.BeTrueV(contexts.MqttBrokerFromContext(ctx))
	l := must.BeTrueV(contexts.LoggerFromContext(ctx))
	cli, err := mq.NewClient(uuid.NewString(), topic)
	if err != nil {
		return err
	}
	defer mq.Close(cli)

	var data any
	switch data.(type) {
	case string:
		data = v
	case []byte:
		data = v
	default:
		data, err = json.Marshal(v)
		if err != nil {
			return err
		}
	}
	if err = cli.Publish(data); err != nil {
		err = errors.Wrap(err, "failed to publish mqtt")
		l.Error(err, "client_id", id, "topic", topic, "data", v)
		return err
	}
	return nil
}
