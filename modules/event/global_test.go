package event_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotexproject/pebble-server/modules/event"
)

func TestNewEvent(t *testing.T) {
	r := require.New(t)
	r.Nil(event.NewEvent(""))

	for _, v := range event.Events() {
		if vv, ok := v.(event.EventHasBlockchainMeta); ok {
			t.Log(v.Source(), vv.ContractID(), vv.EventName(), vv.Topic())
		}
	}
	for _, v := range event.Events() {
		if _, ok := v.(event.EventHasTopicData); ok {
			t.Log(v.Source(), v.Topic())
		}
	}
}
