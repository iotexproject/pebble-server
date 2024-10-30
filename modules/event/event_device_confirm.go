package event

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/xoctopus/x/misc/must"
	"google.golang.org/protobuf/proto"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/iotexproject/pebble-server/contexts"
	"github.com/iotexproject/pebble-server/enums"
	"github.com/iotexproject/pebble-server/models"
	"github.com/iotexproject/pebble-server/pebblepb"
)

func init() {
	e := &DeviceConfirm{}
	registry(e.Topic(), func() Event { return &DeviceConfirm{} })
}

type DeviceConfirm struct {
	IMEI
	SignatureValidator
	pkg *pebblepb.ConfirmPackage
}

func (e *DeviceConfirm) Source() enums.EventSourceType {
	return enums.EVENT_SOURCE_TYPE__MQTT
}

func (e *DeviceConfirm) Topic() string { return "device/+/confirm" }

func (e *DeviceConfirm) UnmarshalTopic(topic []byte) error {
	return (&TopicParser{e, topic, "device", "confirm"}).Unmarshal()
}

func (e *DeviceConfirm) Unmarshal(v any) (err error) {
	data, ok := v.([]byte)
	must.BeTrueWrap(ok, "assertion unmarshal with bytes")

	defer func() { err = WrapUnmarshalError(err, e) }()

	pkg := &pebblepb.ConfirmPackage{}
	if err = proto.Unmarshal(data, pkg); err != nil {
		return errors.Wrap(err, "failed to unmarshal proto")
	}
	e.pkg = pkg

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

type message struct {
	IMEI        string `json:"imei"`
	Owner       string `json:"owner"`
	Timestamp   uint32 `json:"timestamp"`
	Signature   string `json:"signature"`
	GasLimit    string `json:"gasLimit"`
	DataChannel uint32 `json:"dataChannel"`
}

func (e *DeviceConfirm) Handle(ctx context.Context) (err error) {
	defer func() {
		err = WrapHandleError(err, e)
	}()

	// if !contexts.IMEIFilter().MustFrom(ctx).NeedHandle(e.Imei) {
	// 	return errors.Errorf("imei %s not in whitelist", e.Imei)
	// }

	dev := &models.Device{ID: e.Imei}
	err = FetchByPrimary(ctx, dev)
	if err != nil {
		return errors.Wrapf(err, "failed to fetch dev: %s", dev.ID)
	}

	if dev.Status != models.PROPOSAL {
		return errors.Errorf("device `%s` is %d, donnot need confirm", dev.ID, dev.Status)
	}

	e.addr = common.HexToAddress(dev.Address)
	if !e.Validate() {
		return WrapValidateError(e)
	}

	id := uuid.NewString()
	projectID := contexts.ProjectID().MustFrom(ctx)
	projectVersion := contexts.ProjectVersion().MustFrom(ctx)
	msg := &models.Message{
		MessageID:      dev.Address + fmt.Sprintf("-%d-%s", e.pkg.GetTimestamp(), id),
		ClientID:       dev.Address,
		ProjectID:      projectID,
		ProjectVersion: projectVersion,
		Data: must.NoErrorV(json.Marshal(message{
			IMEI:        e.Imei,
			Owner:       common.BytesToAddress(e.pkg.GetOwner()).String(),
			Timestamp:   e.pkg.GetTimestamp(),
			Signature:   hex.EncodeToString(e.sig),
			GasLimit:    big.NewInt(200000).String(),
			DataChannel: uint32(dev.DataChannel),
		})),
		InternalTaskID: id,
	}
	task := &models.Task{
		ProjectID:      projectID,
		InternalTaskID: id,
		MessageIDs:     datatypes.JSON([]byte(`["` + msg.MessageID + `"]`)),
		Signature:      "",
	}

	sk := contexts.PrivateKey().MustFrom(ctx)
	db := contexts.Database().MustFrom(ctx)
	err = db.Transaction(func(tx *gorm.DB) error {
		if err = tx.Create(msg).Error; err != nil {
			return errors.Wrapf(err, "failed to create message")
		}
		if err = tx.Create(task).Error; err != nil {
			return errors.Wrapf(err, "failed to create task")
		}
		if err = task.Sign(sk.PrivateKey, msg); err != nil {
			return errors.Wrapf(err, "failed to sign message")
		}
		err = tx.Model(task).Update("signature", task.Signature).
			Where("id=?", task.ID).Error
		return errors.Wrapf(err, "failed to update task signature")
	})
	return errors.Wrap(err, "in transaction")
}
