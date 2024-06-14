package event

import "context"

func init() {
	e := &PebbleConfig{}
	registry(e.Topic(), func() Event { return &PebbleConfig{} })
}

type PebbleConfig struct {
	imei  string
	appid string
}

func (e *PebbleConfig) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *PebbleConfig) Topic() string {
	return "Config(string imei, string config)"
}

func (e *PebbleConfig) Unmarshal(v any) error {
	// unmarshal event log
	return nil
}

func (e *PebbleConfig) Handle(ctx context.Context) error {
	// update device set config = $appid where id = $imei
	// appv2 := select * from app_v2 where id = $appid
	// if appv2 is not exist, return nil
	// notify device config updated
	// payload: appv2.Data
	// topic: backend/$imei/config
	return nil
}
