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
	e := &DeviceConfirm{}
	registry(e.Topic(), func() Event { return &DeviceConfirm{} })
}

type DeviceConfirm struct {
	imei string
	hash []byte
	sig  []byte
}

func (e *DeviceConfirm) Source() SourceType {
	return SourceTypeMQTT
}

func (e *DeviceConfirm) Topic() string {
	return "device/+/confirm"
}

func (e *DeviceConfirm) Unmarshal(v any) error {
	data, ok := v.([]byte)
	must.BeTrueWrap(ok, "assertion unmarshal with bytes")

	pkg := &pebblepb.ConfirmPackage{}
	if err := proto.Unmarshal(data, pkg); err != nil {
		return err
	}

	var (
		sig   = pkg.GetSignature()
		owner = pkg.GetOwner()
		ts    = pkg.GetTimestamp()
	)

	if len(sig) != 64 {
		return &UnmarshalError{}
	}
	e.sig = append(sig, 0)

	buf := make([]byte, len(owner)+4)
	copy(buf, owner)
	gByteOrder.PutUint32(buf[len(owner):], ts)
	sum := sha256.Sum256(buf)

	e.hash = sum[:]

	return nil
}

func (e *DeviceConfirm) UnmarshalTopic(topic []byte) error {
	parts := bytes.Split(topic, []byte("/"))
	if len(parts) != 3 {
		return &UnmarshalTopicError{}
	}
	if !bytes.Equal(parts[0], []byte("device")) ||
		!bytes.Equal(parts[2], []byte("confirm")) {
		return &UnmarshalTopicError{}
	}
	if len(parts[1]) == 0 {
		return &UnmarshalTopicError{}
	}
	e.imei = string(parts[1])
	return nil
}

func (e *DeviceConfirm) Handle(ctx context.Context) error {
	dev := &models.Device{}

	if dev.Status != 1 {
		// ignored device status
		return nil
	}
	if !ValidateSignature(e.hash, e.sig, address.HexToAddress(dev.Address)) {
		return &ValidateError{}
	}

	// todo bc confirm
	return nil
}
