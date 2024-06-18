package event

import (
	"context"
	"strings"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
)

func init() {
	f := func() Event { return &PebbleConfig{} }
	e := f()
	registry(e.Topic(), f)
}

type PebbleConfig struct {
	Imei   string
	Config string
}

func (e *PebbleConfig) Source() SourceType { return SOURCE_TYPE__BLOCKCHAIN }

func (e *PebbleConfig) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *PebbleConfig) ContractID() string { return enums.CONTRACT__PEBBLE_DEVICE }

func (e *PebbleConfig) EventName() string { return "Config" }

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
