package event_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xoctopus/x/misc/must"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/event"
)

func TestDeviceDataUnmarshal(t *testing.T) {
	r := require.New(t)

	v, ok := event.NewEvent("device/+/data").(*event.DeviceData)
	r.NotNil(v)
	r.True(ok)

	b64s := []string{
		"CAASRQjIARCzAxiA0KzzDiCA0KzzDiiw08ABMIQtOLmDBkCHHEgAUMIWWgMIDQxiBjPOAcKDAWoQYWEyMzNjODRiY2MwNzk0MBiemcCzBiJATJa13Bb09C6pfOEXwHBibqVXtY9Nm0MBiqYhBZsS+3cLqPEec4jPxUBDsfK6DehS2vf050PValjAyPVxnqoy8A==",
	}

	for _, b64 := range b64s {
		raw := must.NoErrorV(base64.StdEncoding.DecodeString(b64))
		r.NoError(v.Unmarshal(raw))
	}

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
}
