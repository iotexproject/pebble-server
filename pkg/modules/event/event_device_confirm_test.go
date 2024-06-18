package event_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/event"
)

func TestDeviceConfirmUnmarshal(t *testing.T) {
	r := require.New(t)

	t.Run("UnmarshalTopic", func(t *testing.T) {
		v := &event.DeviceConfirm{}
		t.Run("Invalid", func(t *testing.T) {
			for _, topic := range [][]byte{
				[]byte("device/abc/data"),
				[]byte("device//confirm/"),
				[]byte("backend/def/confirm"),
				[]byte("device/abc/data"),
			} {
				err := v.UnmarshalTopic(topic)
				r.Error(err)
				r.IsType(err, &event.UnmarshalTopicError{})
			}
		})
		t.Run("Valid", func(t *testing.T) {
			for _, c := range []*struct {
				topic []byte
				imei  string
			}{
				{[]byte("device/abc/confirm"), "abc"},
				{[]byte("device/def/confirm"), "def"},
			} {
				r.NoError(v.UnmarshalTopic(c.topic))
				r.Equal(v.GetIMEI(), c.imei)
			}
		})
	})
}
