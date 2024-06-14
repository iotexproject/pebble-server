package event

import (
	"bytes"
	"context"
)

func init() {
	e := &DeviceQuery{}
	registry(e.Topic(), func() Event { return &DeviceQuery{} })
}

type DeviceQuery struct {
	imei string
}

func (e *DeviceQuery) Source() SourceType {
	return SourceTypeMQTT
}

func (e *DeviceQuery) Topic() string {
	return "device/+/query"
}

func (e *DeviceQuery) Unmarshal(_ any) error {
	return nil
}

func (e *DeviceQuery) UnmarshalTopic(topic []byte) error {
	parts := bytes.Split(topic, []byte("/"))
	if len(parts) != 3 {
		return &UnmarshalTopicError{}
	}
	if !bytes.Equal(parts[0], []byte("device")) ||
		!bytes.Equal(parts[2], []byte("query")) {
		return &UnmarshalTopicError{}
	}
	if len(parts[1]) == 0 {
		return &UnmarshalTopicError{}
	}
	e.imei = string(parts[1])
	return nil
}

type StateQueryResult struct {
	Status   int    `json:"status"`
	Proposer string `json:"proposer,omitempty"`
	Firmware string `json:"firmware,omitempty"`
	URI      string `json:"uri,omitempty"`
	Version  string `json:"version,omitempty"`
}

func (e *DeviceQuery) Handle(ctx context.Context) error {
	// fetch device by imei from `device`
	// fetch firmwares from `app`
	// response `StateQueryResult` to "backend/$imei/status"
	return nil
}
