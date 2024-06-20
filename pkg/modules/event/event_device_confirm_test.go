package event_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/event"
)

func TestDeviceConfirmUnmarshal(t *testing.T) {
	t.Skip("need postgres dependency")
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

	b64 := "ChSL8XCgJ0rpBriLIjTslUibYN6lfhCyvMKzBhpA01gdM1K0JG72+ZgLvA09MUNVEIoMKiaDuK/pbmUt3P8VW70XAyiVOy4Fstz6GotAYiilAPaGuSYg8HA+IBZOSyD3Pw=="
	raw, err := base64.StdEncoding.DecodeString(b64)
	r.NoError(err)
	ctx := testctx()
	v := &event.DeviceConfirm{}

	r.NoError(v.UnmarshalTopic([]byte("device/351358815439952/confirm")))

	r.NoError(v.Unmarshal(raw))

	r.NoError(v.Handle(ctx))
}
