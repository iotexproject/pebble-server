package event

import (
	"context"
	"crypto/sha256"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/xoctopus/x/misc/must"
	"google.golang.org/protobuf/proto"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/pebblepb"
)

func init() {
	e := &DeviceData{}
	registry(e.Topic(), func() Event { return &DeviceData{} })
}

type DeviceData struct {
	IMEI
	SignatureValidator
	pkg proto.Message
}

func (e *DeviceData) Source() SourceType { return SOURCE_TYPE__MQTT }

func (e *DeviceData) Topic() string { return "device/+/data" }

func (e *DeviceData) Unmarshal(v any) (err error) {
	data, ok := v.([]byte)
	must.BeTrueWrap(ok, "assertion unmarshal with bytes")

	defer func() { err = WrapUnmarshalError(err, e) }()

	pkg := &pebblepb.BinPackage{}
	if err = proto.Unmarshal(data, pkg); err != nil {
		return errors.Wrap(err, "failed to unmarshal proto")
	}

	var (
		typ = uint32(pkg.GetType())
		pl  = pkg.GetData()
		ts  = pkg.GetTimestamp()
		sig = pkg.GetSignature()
	)
	if len(sig) != 64 {
		return errors.Errorf("unexpected sig, expect 64 bytes but got %d", len(sig))
	}
	e.sig = append(sig, 0)

	switch t := pkg.GetType(); t {
	case pebblepb.BinPackage_CONFIG:
		e.pkg = &pebblepb.SensorConfig{}
	case pebblepb.BinPackage_STATE:
		e.pkg = &pebblepb.SensorState{}
	case pebblepb.BinPackage_DATA:
		e.pkg = &pebblepb.SensorData{}
	default:
		return errors.Errorf("unexpected senser package type: %d", t)
	}

	if err = proto.Unmarshal(pl, e.pkg); err != nil {
		return errors.Wrapf(err, "failed to unmarshal senser package")
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
	return (&TopicUnmarshaller{e, topic, "device", "data"}).Unmarshal()
}

func (e *DeviceData) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	dev, err := FetchDeviceByIMEI(ctx, e.imei)
	if err != nil {
		return err
	}

	e.addr = common.HexToAddress(dev.Address)
	if !e.Validate() {
		return WrapValidateError(e)
	}

	switch pkg := e.pkg.(type) {
	case *pebblepb.SensorConfig:
		return e.handleConfig(ctx, pkg)
	case *pebblepb.SensorState:
		return e.handleState(ctx, pkg)
	case *pebblepb.SensorData:
		return e.handleData(ctx, pkg)
	default:
		return errors.Errorf("unexpected senser package type: %d", pkg)
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
