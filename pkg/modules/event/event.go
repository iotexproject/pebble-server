package event

import (
	"context"
	"sort"

	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
)

// SourceType defines event source types
type SourceType uint8

const (
	SourceTypeMQTT = iota + 1
	SourceTypeBlockchain
)

type Event interface {
	// Source returns event source type
	Source() SourceType
	Topic() string
	Unmarshal(data []byte) error
	Handle(ctx context.Context) error
}

// EventHasTopicData if an event has required data in topic
type EventHasTopicData interface {
	UnmarshalTopic(topic []byte) error
}

var gEventFactory map[string]func() Event

type factory func() Event

func registry(topic string, f factory) {
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
