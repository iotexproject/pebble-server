package event_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/event"
)

func TestDeviceQuery_Handle(t *testing.T) {
	t.Skip("need postgres dependency")
	r := require.New(t)

	v, ok := event.NewEvent("device/+/query").(*event.DeviceQuery)
	r.NotNil(v)
	r.True(ok)

	topic := []byte("device/350916067070535/query")

	r.NoError(v.UnmarshalTopic(topic))

	r.NoError(v.Unmarshal([]byte("")))

	r.NoError(v.Handle(testctx()))
}
