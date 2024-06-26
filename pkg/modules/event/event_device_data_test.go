package event_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/event"
)

func TestDeviceDataUnmarshal(t *testing.T) {
	// t.Skip("need postgres dependency")
	r := require.New(t)

	v, ok := event.NewEvent("device/+/data").(*event.DeviceData)
	r.NotNil(v)
	r.True(ok)

	t.Run("UnmarshalTopic", func(t *testing.T) {
		t.Run("Invalid", func(t *testing.T) {
			for _, topic := range [][]byte{
				[]byte("/device/abc/data"),
				[]byte("device//data/"),
				[]byte("backend/def/confirm"),
				[]byte("device/abc/confirm"),
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
				{[]byte("device/abc/data"), "abc"},
				{[]byte("device/def/data"), "def"},
			} {
				r.NoError(v.UnmarshalTopic(c.topic))
				r.Equal(v.GetIMEI(), c.imei)
			}
		})
	})

	b64 := "CAASRAjMCBDkAhiAwIncAyC/kfCeByj0l6ABMLogOID+BUD+JUgAUJ0QWgMsDwZiBTXmf+ICahA1ZGEwOTk3NDE0OTIyYThiGLbUyLMGIkC4zqnDST/KDHLihv8Nbks5gu6/XZ8kB/ytvF0YUA0bQoL7QbuPbmNeROcBKNj4ePkKihnjI373E8NaybO0wmte"
	raw, err := base64.StdEncoding.DecodeString(b64)
	r.NoError(err)

	topic := []byte("device/350916067070535/data")

	r.NoError(v.UnmarshalTopic(topic))

	r.NoError(v.Unmarshal(raw))

	r.NoError(v.Handle(testctx()))

}
