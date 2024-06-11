package event

import "context"

func init() {
	e := &PebbleFirmware{}
	registry(e.Topic(), func() Event { return &PebbleFirmware{} })
}

type PebbleFirmware struct {
	imei  string
	appid string
}

func (e *PebbleFirmware) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *PebbleFirmware) Topic() string {
	return "Firmware(string imei, string app)"
}

func (e *PebbleFirmware) Unmarshal(data []byte) error {
	// unmarshal event log
	return nil
}

func (e *PebbleFirmware) Handle(ctx context.Context) error {
	// app := select * from app where id = $appid
	// if app is not exist, return err
	// update device set firmware = '$app.id app.version' where id = $imei
	// notify device firmware updated
	// payload {firmware: $appid, uri: app.uri, version: app.version}
	// topic: backend/$imei/firmware
	return nil
}
