package event

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"github.com/xoctopus/x/misc/must"
	"google.golang.org/protobuf/proto"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
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
	bin *pebblepb.BinPackage
}

func (e *DeviceData) Source() enums.EventSourceType {
	return enums.EVENT_SOURCE_TYPE__MQTT
}

func (e *DeviceData) Topic() string { return "device/+/data" }

func (e *DeviceData) Unmarshal(v any) (err error) {
	data, ok := v.([]byte)
	must.BeTrueWrap(ok, "assertion unmarshal with bytes")

	defer func() { err = WrapUnmarshalError(err, e) }()

	pkg := &pebblepb.BinPackage{}
	if err = proto.Unmarshal(data, pkg); err != nil {
		return errors.Wrap(err, "failed to unmarshal proto")
	}
	e.bin = pkg

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
	return (&TopicParser{e, topic, "device", "data"}).Unmarshal()
}

func (e *DeviceData) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	if !contexts.IMEIFilter().MustFrom(ctx).NeedHandle(e.Imei) {
		return errors.Errorf("imei %s not in whitelist", e.Imei)
	}

	dev := &models.Device{ID: e.Imei}
	if err = FetchByPrimary(ctx, dev); err != nil {
		return errors.Wrapf(err, "failed to fetch dev: %s", dev.ID)
	}

	// TODO need ?
	// if dev.Status != models.CONFIRM {
	// 	return errors.Errorf("require device confirmed(2), but got %d", dev.Status)
	// }

	e.addr = common.HexToAddress(dev.Address)
	if !e.Validate() {
		return WrapValidateError(e)
	}

	switch pkg := e.pkg.(type) {
	case *pebblepb.SensorConfig:
		err = e.handleConfig(ctx, dev, pkg)
	case *pebblepb.SensorState:
		err = e.handleState(ctx, dev, pkg)
	case *pebblepb.SensorData:
		err = e.handleSensor(ctx, dev, pkg)
	default:
		err = errors.Errorf("unexpected senser package type")
	}
	return errors.Wrapf(err, "failed to handle %T", e.pkg)
}

func (e *DeviceData) handleConfig(ctx context.Context, dev *models.Device, pkg *pebblepb.SensorConfig) error {
	dev.BulkUpload = int32(pkg.GetBulkUpload())
	dev.DataChannel = int32(pkg.GetDataChannel())
	dev.UploadPeriod = int32(pkg.GetUploadPeriod())
	dev.BulkUploadSamplingCnt = int32(pkg.GetBulkUploadSamplingCnt())
	dev.BulkUploadSamplingFreq = int32(pkg.GetBulkUploadSamplingFreq())
	dev.Beep = int32(pkg.GetBeep())
	dev.RealFirmware = pkg.GetFirmware()
	dev.Configurable = pkg.GetDeviceConfigurable()
	dev.UpdatedAt = time.Now()

	err := UpdateByPrimary(ctx, dev, map[string]any{
		"bulk_upload":               dev.BulkUpload,
		"data_channel":              dev.DataChannel,
		"upload_period":             dev.UploadPeriod,
		"bulk_upload_sampling_cnt":  dev.BulkUploadSamplingCnt,
		"bulk_upload_sampling_freq": dev.BulkUploadSamplingFreq,
		"beep":                      dev.Beep,
		"real_firmware":             dev.RealFirmware,
		"configurable":              dev.Configurable,
		"updated_at":                dev.UpdatedAt,
	})
	return errors.Wrapf(err, "failed to upsert device config: %s", dev.ID)
}

func (e *DeviceData) handleState(ctx context.Context, dev *models.Device, pkg *pebblepb.SensorState) error {
	dev.State = int32(pkg.GetState())

	err := UpdateByPrimary(ctx, dev, map[string]any{
		"state":      dev.State,
		"updated_at": time.Now(),
	})
	return errors.Wrapf(err, "failed to update device state: %s %d", dev.ID, dev.State)
}

func (e *DeviceData) handleSensor(ctx context.Context, dev *models.Device, pkg *pebblepb.SensorData) error {
	snr := float64(pkg.GetSnr())
	if snr > 2700 {
		snr = 100
	} else if snr < 700 {
		snr = 25
	} else {
		snr, _ = big.NewFloat((snr-700)*0.0375 + 25).Float64()
	}

	vbat := (float64(pkg.GetVbat()) - 320) / 90
	if vbat > 1 {
		vbat = 100
	} else if vbat < 0.1 {
		vbat = 0.1
	} else {
		vbat *= 100
	}

	gyroscope, _ := json.Marshal(pkg.GetGyroscope())
	accelerometer, _ := json.Marshal(pkg.GetAccelerometer())

	dr := &models.DeviceRecord{
		ID:             dev.ID + "-" + fmt.Sprintf("%d", e.bin.GetTimestamp()),
		Imei:           dev.ID,
		Timestamp:      int64(e.bin.GetTimestamp()),
		Signature:      hex.EncodeToString(e.sig),
		Operator:       "",
		Snr:            strconv.FormatFloat(snr, 'f', 1, 64),
		Vbat:           strconv.FormatFloat(vbat, 'f', 1, 64),
		Latitude:       decimal.NewFromInt32(pkg.GetLatitude()).Div(decimal.NewFromInt32(10000000)).StringFixed(7),
		Longitude:      decimal.NewFromInt32(pkg.GetLongitude()).Div(decimal.NewFromInt32(10000000)).StringFixed(7),
		GasResistance:  decimal.NewFromInt32(int32(pkg.GetGasResistance())).Div(decimal.NewFromInt32(100)).StringFixed(2),
		Temperature:    decimal.NewFromInt32(pkg.GetTemperature()).Div(decimal.NewFromInt32(100)).StringFixed(2),
		Temperature2:   decimal.NewFromInt32(int32(pkg.GetTemperature2())).Div(decimal.NewFromInt32(100)).StringFixed(2),
		Pressure:       decimal.NewFromInt32(int32(pkg.GetPressure())).Div(decimal.NewFromInt32(100)).StringFixed(2),
		Humidity:       decimal.NewFromInt32(int32(pkg.GetHumidity())).Div(decimal.NewFromInt32(100)).StringFixed(2),
		Light:          decimal.NewFromInt32(int32(pkg.GetLight())).Div(decimal.NewFromInt32(100)).StringFixed(2),
		Gyroscope:      string(gyroscope),
		Accelerometer:  string(accelerometer),
		OperationTimes: models.NewOperationTimes(),
	}
	submit(ctx, dr)
	_, err := UpsertOnConflict(ctx, dr, "id")
	return errors.Wrapf(err, "failed to upsert senser data: %s", dev.ID)
}
