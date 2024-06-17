package event

import (
	"bytes"
	"context"
	"crypto/sha256"

	"github.com/xoctopus/x/misc/must"
	"google.golang.org/protobuf/proto"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/ethutil/address"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/pebblepb"
)

func init() {
	e := &DeviceData{}
	registry(e.Topic(), func() Event { return &DeviceData{} })
}

type DeviceData struct {
	imei string
	pkg  proto.Message
	hash []byte
	sig  []byte
}

func (e *DeviceData) Source() SourceType {
	return SourceTypeMQTT
}

func (e *DeviceData) Topic() string {
	return "device/+/data"
}

func (e *DeviceData) Unmarshal(v any) error {
	data, ok := v.([]byte)
	must.BeTrueWrap(ok, "assertion unmarshal with bytes")

	pkg := &pebblepb.BinPackage{}
	if err := proto.Unmarshal(data, pkg); err != nil {
		return &UnmarshalError{}
	}

	var (
		typ = uint32(pkg.GetType())
		pl  = pkg.GetData()
		ts  = pkg.GetTimestamp()
		sig = pkg.GetSignature()
	)
	if len(sig) != 64 {
		return &UnmarshalError{}
	}
	e.sig = append(sig, 0)

	switch pkg.GetType() {
	case pebblepb.BinPackage_CONFIG:
		e.pkg = &pebblepb.SensorConfig{}
	case pebblepb.BinPackage_STATE:
		e.pkg = &pebblepb.SensorState{}
	case pebblepb.BinPackage_DATA:
		e.pkg = &pebblepb.SensorData{}
	default:
		return &UnmarshalError{}
	}
	if err := proto.Unmarshal(pl, e.pkg); err != nil {
		return &UnmarshalError{}
	}

	buf := make([]byte, 4+len(pl)+4)
	gByteOrder.PutUint32(buf, typ)
	copy(buf[4:], pl)
	gByteOrder.PutUint32(buf[4+len(pl):], ts)
	sum := sha256.Sum256(buf)

	e.hash = sum[:]

	return nil
}

func (e *DeviceData) UnmarshalTopic(topic []byte) error {
	parts := bytes.Split(topic, []byte("/"))
	if len(parts) != 3 {
		return &UnmarshalTopicError{}
	}
	if !bytes.Equal(parts[0], []byte("device")) ||
		!bytes.Equal(parts[2], []byte("data")) {
		return &UnmarshalTopicError{}
	}
	if len(parts[1]) == 0 {
		return &UnmarshalTopicError{}
	}
	e.imei = string(parts[1])
	return nil
}

func (e *DeviceData) Handle(ctx context.Context) error {
	device := &models.Device{}

	if !ValidateSignature(e.hash, e.sig, address.HexToAddress(device.Address)) {
		return &ValidateError{}
	}

	switch pkg := e.pkg.(type) {
	case *pebblepb.SensorConfig:
		return e.handleConfig(ctx, pkg)
	case *pebblepb.SensorState:
		return e.handleState(ctx, pkg)
	case *pebblepb.SensorData:
		return e.handleData(ctx, pkg)
	default:
		return &HandleError{}
	}
}

func (e *DeviceData) handleConfig(ctx context.Context, pkg *pebblepb.SensorConfig) error {
	return nil
}

func (e *DeviceData) handleState(ctx context.Context, pkg *pebblepb.SensorState) error {
	return nil
}

func (e *DeviceData) handleData(ctx context.Context, pkg *pebblepb.SensorData) error {
	return nil
}
