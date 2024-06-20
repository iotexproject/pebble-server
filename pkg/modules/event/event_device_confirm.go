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

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
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
	pkg *pebblepb.ConfirmPackage
}

func (e *DeviceConfirm) Source() SourceType { return SOURCE_TYPE__MQTT }

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
		return
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
	DataChannel int32  `json:"dataChannel"`
}

func (e *DeviceConfirm) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	dev := &models.Device{ID: e.imei}
	err = FetchByPrimary(ctx, dev)
	if err != nil {
		return err
	}

	if dev.Status != models.PROPOSAL {
		return errors.Errorf("device `%s` is %d, donnt need confirm", dev.ID, dev.Status)
	}

	e.addr = common.HexToAddress(dev.Address)
	if !e.Validate() {
		return WrapValidateError(e)
	}

	id := uuid.NewString()
	msg := &models.Message{
		MessageID:      dev.Address + fmt.Sprintf("-%d", e.pkg.GetTimestamp()),
		ClientID:       dev.Address,
		ProjectID:      must.BeTrueV(contexts.ProjectIDFromContext(ctx)),
		ProjectVersion: must.BeTrueV(contexts.ProjectVersionFromContext(ctx)),
		Data: must.NoErrorV(json.Marshal([]message{{
			IMEI:        e.imei,
			Owner:       dev.Owner,
			Timestamp:   e.pkg.GetTimestamp(),
			Signature:   hex.EncodeToString(e.pkg.GetSignature()),
			GasLimit:    big.NewInt(200000).String(),
			DataChannel: dev.DataChannel,
		}})),
		InternalTaskID: id,
	}
	task := &models.Task{
		ProjectID:      must.BeTrueV(contexts.ProjectIDFromContext(ctx)),
		InternalTaskID: id,
		MessageIDs:     datatypes.JSON([]byte(`[` + id + `]`)),
		Signature:      "",
	}

	sk := must.BeTrueV(contexts.EcdsaPrivateKeyFromContext(ctx))
	db, _ := contexts.DatabaseFromContext(ctx)
	return db.Transaction(func(tx *gorm.DB) error {
		if err = tx.Create(msg).Error; err != nil {
			return err
		}
		if err = tx.Create(task).Error; err != nil {
			return err
		}
		if err = task.Sign(sk, msg); err != nil {
			return err
		}
		return tx.Model(task).
			Update("signature", task.Signature).Where("id=?", task.ID).Error
	})
}
