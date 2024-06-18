package event

import (
	"context"
	"crypto/sha256"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/xoctopus/x/misc/must"
	"google.golang.org/protobuf/proto"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/pebblepb"
)

func init() {
	e := &DeviceConfirm{}
	registry(e.Topic(), func() Event { return &DeviceConfirm{} })
}

type DeviceConfirm struct {
	IMEI
	SignatureValidator
}

func (e *DeviceConfirm) Source() SourceType { return SOURCE_TYPE__MQTT }

func (e *DeviceConfirm) Topic() string { return "device/+/confirm" }

func (e *DeviceConfirm) UnmarshalTopic(topic []byte) error {
	return (&TopicUnmarshaller{e, topic, "device", "confirm"}).Unmarshal()
}

func (e *DeviceConfirm) Unmarshal(v any) (err error) {
	data, ok := v.([]byte)
	must.BeTrueWrap(ok, "assertion unmarshal with bytes")

	defer func() { err = WrapUnmarshalError(err, e) }()

	pkg := &pebblepb.ConfirmPackage{}
	if err = proto.Unmarshal(data, pkg); err != nil {
		return
	}

	var (
		sig   = pkg.GetSignature()
		owner = pkg.GetOwner()
		ts    = pkg.GetTimestamp()
	)

	if len(sig) != 64 {
		return errors.Errorf("unexpected sig, expect 64 bytes but got %d", len(sig))
	}
	e.sig = append(sig, 0)

	buf := make([]byte, len(owner)+4)
	copy(buf, owner)
	gByteOrder.PutUint32(buf[len(owner):], ts)

	sum := sha256.Sum256(buf)
	e.hash = sum[:]

	return nil
}

func (e *DeviceConfirm) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	dev, err := FetchDeviceByIMEI(ctx, e.imei)
	if err != nil {
		return
	}

	if dev.Status != int32(models.PROPOSAL) {
		return errors.Errorf("device `%s` is %d, donnt need confirm", dev.ID, dev.Status)
	}

	e.addr = common.HexToAddress(dev.Address)
	if !e.Validate() {
		return WrapValidateError(e)
	}

	// todo commit blockchain task to prover
	return nil
}
