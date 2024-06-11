package event_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/event"
)

func TestNewEvent(t *testing.T) {
	r := require.New(t)
	r.Nil(event.NewEvent(""))

	for _, topic := range event.Topics() {
		t.Log(topic)
	}

	for _, v := range event.Events() {
		t.Log(v.Source(), v.Topic())
	}
}
