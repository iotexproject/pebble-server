package event

import (
	"context"
	"fmt"
	"sort"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"github.com/xoctopus/x/misc/must"
	"golang.org/x/exp/maps"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/blockchain"
)

// SourceType defines event source types
type SourceType uint8

const (
	SourceTypeMQTT = iota + 1
	SourceTypeBlockchain
)

type Event interface {
	// Source returns event source type, eg: mqtt, blockchain
	Source() SourceType
	// Topic returns the event subscription topic
	Topic() string
	// Unmarshal used to parse message payload to event body
	Unmarshal(data any) error
	// Handle to process event
	Handle(ctx context.Context) error
}

// EventHasTopicData if an event has metadata in topic, such as device imei in topic
type EventHasTopicData interface {
	UnmarshalTopic(topic []byte) error
}

var gEventFactory map[string]func() Event

func registry(topic string, f func() Event) {
	if gEventFactory == nil {
		gEventFactory = make(map[string]func() Event)
	}
	_, ok := gEventFactory[topic]
	if ok {
		panic(errors.Errorf("topic %s reregisted", topic))
	}
	gEventFactory[topic] = f
}

func NewEvent(topic string) Event {
	f, ok := gEventFactory[topic]
	if ok {
		return f()
	}
	return nil
}

func Topics() []string {
	topics := maps.Keys(gEventFactory)
	sort.Slice(topics, func(i, j int) bool {
		return topics[i] < topics[j]
	})
	return topics
}

func Events() []Event {
	events := make([]Event, 0, len(gEventFactory))
	for _, f := range gEventFactory {
		events = append(events, f())
	}
	return events
}

func Handle(ctx context.Context, subtopic, topic string, data any) error {
	factory := gEventFactory[subtopic]
	if factory == nil {
		return errors.Errorf("factory not found by topic: %s", subtopic)
	}
	event := factory()

	if parser, ok := event.(EventHasTopicData); ok {
		if err := parser.UnmarshalTopic([]byte(topic)); err != nil {
			return err
		}
	}
	if err := event.Unmarshal(data); err != nil {
		return err
	}
	return event.Handle(ctx)
}

func InitRunner(ctx context.Context) func() {
	return func() {
		if err := Init(ctx); err != nil {
			fmt.Printf("event module initialize failed: %v\n", err)
			panic(err)
		}
		fmt.Println("event module initialized")
	}
}

func Init(ctx context.Context) error {
	logger, ok := contexts.LoggerFromContext(ctx)
	must.BeTrueWrap(ok, "expect logger from context")
	broker, ok := contexts.MqttBrokerFromContext(ctx)
	must.BeTrueWrap(ok, "expect mqtt broker from context")
	bc, ok := contexts.BlockchainFromContext(ctx)
	must.BeTrueWrap(ok, "expect blockchain from context")

	for _, v := range Events() {
		switch v.Source() {
		case SourceTypeMQTT:
			client, err := broker.NewClient(v.Topic(), v.Topic())
			if err != nil {
				return errors.Wrapf(err, "failed to new mqtt client for topic %s", v.Topic())
			}
			err = client.Subscribe(func(_ mqtt.Client, message mqtt.Message) {
				err := Handle(ctx, v.Topic(), message.Topic(), message.Payload())
				if err != nil {
					logger.WithValues(
						"client", client.ID(),
						"topic", v.Topic(),
					).Error(err, "failed to handle mqtt message")
				}
			})
			if err != nil {
				return errors.Wrapf(err, "failed to subscribe topic: %s", v.Topic())
			}
			return nil
		case SourceTypeBlockchain:
			inst, err := bc.NewMonitorDefault(v.Topic(), v.Topic())
			if err != nil {
				return errors.Wrapf(err, "failed to new monitor instance %s", v.Topic())
			}
			err = inst.Subscribe(func(inst *blockchain.MonitorInstance, message *blockchain.Message) {
				err := Handle(ctx, v.Topic(), message.Topic(), message.Log)
				if err != nil {
					logger.WithValues(
						"instance", inst.ID,
						"topic", v.Topic(),
					).Error(err, "failed to handle mqtt message")
				}
			})
			return nil
		default:
			return errors.Errorf("unexpected event source type: %d", v)
		}
	}
	return nil
}
