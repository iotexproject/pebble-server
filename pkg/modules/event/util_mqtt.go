package event

import (
	"context"
	"encoding/json"

	"github.com/xoctopus/x/misc/must"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
)

func PublicMqttMessage(ctx context.Context, id, topic string, v any) error {
	mq := must.BeTrueV(contexts.MqttBrokerFromContext(ctx))
	cli, err := mq.NewClient(id, topic)
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
	return cli.Publish(data)
}
