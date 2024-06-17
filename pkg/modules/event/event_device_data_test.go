package event_test

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/event"
)

func TestDeviceConfirm_Unmarshal(t *testing.T) {
	r := require.New(t)
	v := event.NewEvent("device/+/data")
	r.NotNil(v)

	raw := "CAASRQjUFhDEAxjA6ZWKAiD/5rmVBiiA5aoxMNMWOI6ZBkDiJEgAUNMWWgMJAQhiBvwOxyf0emoQMWI2NDk0NzRiMTM0ZWNiYxjogb6zBiJAMRDkPDyYDVHt72ZygcxJljB+/vpVf2SV0EzQ5YLmh5kHHEFFYZquQB2bdqMnj/dyWpLKOmORfoGm4kn5s07xpQ=="
	data, err := base64.StdEncoding.DecodeString(raw)
	r.NoError(err)

	r.NoError(v.Unmarshal(data))

	v.Handle(nil)
}
