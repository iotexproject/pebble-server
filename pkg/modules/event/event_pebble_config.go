package event

import (
	"context"
	"strings"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
)

func init() {
	f := func() Event {
		return &PebbleConfig{
			contractID: enums.CONTRACT__PEBBLE_DEVICE,
		}
	}
	e := f()
	registry(e.Topic(), f)
}

type PebbleConfig struct {
	Imei   string
	Config string

	contractID string
}

func (e *PebbleConfig) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *PebbleConfig) Topic() string {
	return network.Topic(e.contractID) + "__" + strings.ToUpper(e.EventName())
}

func (e *PebbleConfig) ContractID() string {
	return network.ContractID(e.contractID)
}

func (e *PebbleConfig) EventName() string {
	return "Config"
}

func (e *PebbleConfig) SubscriberID() string {
	return network.SubscriberID(e.contractID)
}

func (e *PebbleConfig) Data() any { return e }

func (e *PebbleConfig) Unmarshal(any) error { return nil }

func (e *PebbleConfig) Handle(ctx context.Context) error {
	// update device set config = $appid where id = $imei
	// appv2 := select * from app_v2 where id = $appid
	// if appv2 is not exist, return nil
	// notify device config updated
	// payload: appv2.Data
	// topic: backend/$imei/config
	return nil
}
