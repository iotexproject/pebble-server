package event

import "context"

func init() {
	e := &FirmwareRemoved{}
	registry(e.Topic(), func() Event { return &FirmwareRemoved{} })
}

type FirmwareRemoved struct {
	appid string
}

func (e *FirmwareRemoved) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *FirmwareRemoved) Topic() string {
	return "FirmwareRemoved(string name)"
}

func (e *FirmwareRemoved) Unmarshal(v any) error {
	// unmarshal event log
	return nil
}

func (e *FirmwareRemoved) Handle(ctx context.Context) error {
	// remove app by appid
	return nil
}
