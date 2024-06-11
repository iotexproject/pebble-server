package event

import "context"

func init() {
	e := &FirmwareUpdated{}
	registry(e.Topic(), func() Event { return &FirmwareUpdated{} })
}

type FirmwareUpdated struct {
	appid   string
	version string
	uri     string
	avatar  string
}

func (e *FirmwareUpdated) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *FirmwareUpdated) Topic() string {
	return "FirmwareUpdated(string name, string version, string uri, string avatar)"
}

func (e *FirmwareUpdated) Unmarshal(data []byte) error {
	// unmarshal event log
	return nil
}

func (e *FirmwareUpdated) Handle(ctx context.Context) error {
	// create or update app
	// notify device firmware updated device/app_updated/$appid
	// {name:app.id,version:app.version,uri:app.uri,avatar:app.avatar}
	return nil
}
