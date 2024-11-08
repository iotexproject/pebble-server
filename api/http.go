package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shopspring/decimal"
	goproto "google.golang.org/protobuf/proto"

	"github.com/iotexproject/pebble-server/contract/ioid"
	"github.com/iotexproject/pebble-server/contract/ioidregistry"
	"github.com/iotexproject/pebble-server/db"
	"github.com/iotexproject/pebble-server/metrics"
	"github.com/iotexproject/pebble-server/proto"
)

type errResp struct {
	Error string `json:"error,omitempty"`
}

func newErrResp(err error) *errResp {
	return &errResp{Error: err.Error()}
}

type queryReq struct {
	DeviceID  string `json:"deviceID"                   binding:"required"`
	Signature string `json:"signature,omitempty"        binding:"required"`
}

type queryResp struct {
	Status   int32  `json:"status"`
	Owner    string `json:"owner"`
	Firmware string `json:"firmware,omitempty"`
	URI      string `json:"uri,omitempty"`
	Version  string `json:"version,omitempty"`
}

type receiveReq struct {
	DeviceID  string `json:"deviceID"                   binding:"required"`
	Payload   string `json:"payload"                    binding:"required"`
	Signature string `json:"signature,omitempty"        binding:"required"`
}

type httpServer struct {
	engine               *gin.Engine
	db                   *db.DB
	ioidInstance         *ioid.Ioid
	ioidRegistryInstance *ioidregistry.Ioidregistry
}

func (s *httpServer) query(c *gin.Context) {
	req := &queryReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		slog.Error("failed to bind request", "error", err)
		c.JSON(http.StatusBadRequest, newErrResp(errors.Wrap(err, "invalid request payload")))
		return
	}

	deviceAddr := common.HexToAddress(strings.TrimPrefix(req.DeviceID, "did:io:"))
	sigStr := req.Signature
	req.Signature = ""

	ok, err := s.verifySignature(deviceAddr, sigStr, req)
	if err != nil {
		slog.Error("failed to verify signature", "error", err)
		c.JSON(http.StatusBadRequest, newErrResp(errors.Wrap(err, "failed to verify signature")))
		return
	}
	if !ok {
		slog.Error("signature mismatch")
		c.JSON(http.StatusBadRequest, newErrResp(errors.New("signature mismatch")))
		return
	}

	d, err := s.db.Device(req.DeviceID)
	if err != nil {
		slog.Error("failed to query device", "error", err, "device_id", req.DeviceID)
		c.JSON(http.StatusInternalServerError, newErrResp(errors.Wrap(err, "failed to query device")))
		return
	}
	if d == nil {
		nd, code, err := s.ensureDevice(deviceAddr, req.DeviceID)
		if err != nil {
			slog.Error("failed to ensure device", "error", err, "device_id", req.DeviceID)
			c.JSON(code, newErrResp(err))
			return
		}
		d = nd
	}

	metrics.TrackDeviceCount(req.DeviceID)
	metrics.TrackRequestCount("get")
	now := time.Now()
	defer func() {
		metrics.TrackRequestDuration("get", time.Since(now))
	}()

	var (
		firmware string
		uri      string
		version  string
	)
	if parts := strings.Split(d.RealFirmware, " "); len(parts) == 2 {
		app, err := s.db.App(parts[0])
		if err != nil {
			slog.Error("failed to query app", "error", err, "app_id", parts[0])
			c.JSON(http.StatusInternalServerError, newErrResp(errors.Wrap(err, "failed to query app")))
			return
		}
		if app != nil {
			firmware = app.ID
			uri = app.Uri
			version = app.Version
		}
	}

	c.JSON(http.StatusOK, &queryResp{
		Status:   d.Status,
		Owner:    d.Owner,
		Firmware: firmware,
		URI:      uri,
		Version:  version,
	})
}

func (s *httpServer) receive(c *gin.Context) {
	req := &receiveReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		slog.Error("failed to bind request", "error", err)
		c.JSON(http.StatusBadRequest, newErrResp(errors.Wrap(err, "invalid request payload")))
		return
	}

	deviceAddr := common.HexToAddress(strings.TrimPrefix(req.DeviceID, "did:io:"))
	sigStr := req.Signature
	req.Signature = ""

	ok, err := s.verifySignature(deviceAddr, sigStr, req)
	if err != nil {
		slog.Error("failed to verify signature", "error", err)
		c.JSON(http.StatusBadRequest, newErrResp(errors.Wrap(err, "failed to verify signature")))
		return
	}
	if !ok {
		slog.Error("signature mismatch")
		c.JSON(http.StatusBadRequest, newErrResp(errors.New("signature mismatch")))
		return
	}

	d, err := s.db.Device(req.DeviceID)
	if err != nil {
		slog.Error("failed to query device", "error", err, "device_id", req.DeviceID)
		c.JSON(http.StatusInternalServerError, newErrResp(errors.Wrap(err, "failed to query device")))
		return
	}
	if d == nil {
		if _, code, err := s.ensureDevice(deviceAddr, req.DeviceID); err != nil {
			slog.Error("failed to ensure device", "error", err, "device_id", req.DeviceID)
			c.JSON(code, newErrResp(err))
			return
		}
	}

	metrics.TrackDeviceCount(req.DeviceID)
	metrics.TrackRequestCount("post")
	now := time.Now()
	defer func() {
		metrics.TrackRequestDuration("post", time.Since(now))
	}()

	payload, err := base64.RawURLEncoding.DecodeString(req.Payload)
	if err != nil {
		slog.Error("failed to decode base64 data", "error", err)
		c.JSON(http.StatusBadRequest, newErrResp(errors.Wrap(err, "failed to decode base64 data")))
		return
	}
	pkg, data, err := s.unmarshalPayload(payload)
	if err != nil {
		slog.Error("failed to unmarshal payload", "error", err)
		c.JSON(http.StatusBadRequest, newErrResp(errors.Wrap(err, "failed to unmarshal payload")))
		return
	}
	if err := s.handle(req.DeviceID, pkg, data); err != nil {
		slog.Error("failed to handle payload data", "error", err)
		c.JSON(http.StatusInternalServerError, newErrResp(errors.Wrap(err, "failed to handle payload data")))
		return
	}
	c.Status(http.StatusOK)
}

func (s *httpServer) verifySignature(deviceAddr common.Address, sigStr string, o any) (bool, error) {
	reqJson, err := json.Marshal(o)
	if err != nil {
		return false, errors.Wrap(err, "failed to marshal request into json format")
	}
	sig, err := hexutil.Decode(sigStr)
	if err != nil {
		return false, errors.Wrapf(err, "failed to decode signature from hex format, signature %s", sigStr)
	}
	hash := sha256.New()
	hash.Write(reqJson)
	h := hash.Sum(nil)

	res := []common.Address{}
	rID := []uint8{0, 1}
	for _, id := range rID {
		ns := append(sig, byte(id))
		if a, err := s.recover(ns, h); err != nil {
			slog.Info("failed to recover address from signature", "error", err, "recover_id", id, "signature", sigStr)
		} else {
			res = append(res, a)
		}
	}

	for _, r := range res {
		if bytes.Equal(r.Bytes(), deviceAddr.Bytes()) {
			return true, nil
		}
	}
	return false, nil
}

func (s *httpServer) recover(sig, h []byte) (common.Address, error) {
	sigpk, err := crypto.SigToPub(h, sig)
	if err != nil {
		return common.Address{}, errors.Wrapf(err, "failed to recover public key from signature")
	}
	slog.Info("public key", "data", hexutil.Encode(crypto.FromECDSAPub(sigpk)))
	return crypto.PubkeyToAddress(*sigpk), nil
}

func (s *httpServer) ensureDevice(deviceAddr common.Address, deviceID string) (*db.Device, int, error) {
	tokenID, err := s.ioidRegistryInstance.DeviceTokenId(nil, deviceAddr)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "failed to query device token id")
	}
	deviceOwner, err := s.ioidInstance.OwnerOf(nil, tokenID)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "failed to query device owner")
	}

	dev := &db.Device{
		ID:             deviceID,
		Owner:          deviceOwner.String(),
		Address:        deviceAddr.String(),
		Status:         db.CONFIRM,
		Proposer:       deviceOwner.String(),
		OperationTimes: db.NewOperationTimes(),
	}
	if err := s.db.UpsertDevice(dev); err != nil {
		return nil, http.StatusInternalServerError, errors.Wrap(err, "failed to upsert device")
	}
	return dev, http.StatusOK, nil
}

func (s *httpServer) unmarshalPayload(payload []byte) (*proto.BinPackage, goproto.Message, error) {
	pkg := &proto.BinPackage{}
	if err := goproto.Unmarshal(payload, pkg); err != nil {
		return nil, nil, errors.Wrap(err, "failed to unmarshal proto")
	}

	var d goproto.Message
	switch t := pkg.GetType(); t {
	case proto.BinPackage_CONFIG:
		d = &proto.SensorConfig{}
	case proto.BinPackage_STATE:
		d = &proto.SensorState{}
	case proto.BinPackage_DATA:
		d = &proto.SensorData{}
	default:
		return nil, nil, errors.Errorf("unexpected senser package type: %d", t)
	}

	err := goproto.Unmarshal(pkg.GetData(), d)
	return pkg, d, errors.Wrapf(err, "failed to unmarshal senser package")
}

func (s *httpServer) handle(id string, pkg *proto.BinPackage, data goproto.Message) (err error) {
	switch data := data.(type) {
	case *proto.SensorConfig:
		err = s.handleConfig(id, data)
	case *proto.SensorState:
		err = s.handleState(id, data)
	case *proto.SensorData:
		err = s.handleSensor(id, pkg, data)
	}
	return errors.Wrapf(err, "failed to handle %T", data)
}

func (s *httpServer) handleConfig(id string, data *proto.SensorConfig) error {
	err := s.db.UpdateByID(id, map[string]any{
		"bulk_upload":               int32(data.GetBulkUpload()),
		"data_channel":              int32(data.GetDataChannel()),
		"upload_period":             int32(data.GetUploadPeriod()),
		"bulk_upload_sampling_cnt":  int32(data.GetBulkUploadSamplingCnt()),
		"bulk_upload_sampling_freq": int32(data.GetBulkUploadSamplingFreq()),
		"beep":                      int32(data.GetBeep()),
		"real_firmware":             data.GetFirmware(),
		"configurable":              data.GetDeviceConfigurable(),
		"updated_at":                time.Now(),
	})
	return errors.Wrapf(err, "failed to update device config: %s", id)
}

func (s *httpServer) handleState(id string, data *proto.SensorState) error {
	err := s.db.UpdateByID(id, map[string]any{
		"state":      int32(data.GetState()),
		"updated_at": time.Now(),
	})
	return errors.Wrapf(err, "failed to update device state: %s %d", id, int32(data.GetState()))
}

func (s *httpServer) handleSensor(id string, pkg *proto.BinPackage, data *proto.SensorData) error {
	snr := float64(data.GetSnr())
	if snr > 2700 {
		snr = 100
	} else if snr < 700 {
		snr = 25
	} else {
		snr, _ = big.NewFloat((snr-700)*0.0375 + 25).Float64()
	}

	vbat := (float64(data.GetVbat()) - 320) / 90
	if vbat > 1 {
		vbat = 100
	} else if vbat < 0.1 {
		vbat = 0.1
	} else {
		vbat *= 100
	}

	gyroscope, err := json.Marshal(data.GetGyroscope())
	if err != nil {
		errors.Wrap(err, "failed to marshal gyroscope data")
	}
	accelerometer, err := json.Marshal(data.GetAccelerometer())
	if err != nil {
		errors.Wrap(err, "failed to marshal accelerometer data")
	}

	dr := &db.DeviceRecord{
		ID:             id + "-" + fmt.Sprintf("%d", pkg.GetTimestamp()),
		Imei:           id,
		Timestamp:      int64(pkg.GetTimestamp()),
		Signature:      hex.EncodeToString(append(pkg.GetSignature(), 0)),
		Operator:       "",
		Snr:            strconv.FormatFloat(snr, 'f', 1, 64),
		Vbat:           strconv.FormatFloat(vbat, 'f', 1, 64),
		Latitude:       decimal.NewFromInt32(data.GetLatitude()).Div(decimal.NewFromInt32(10000000)).StringFixed(7),
		Longitude:      decimal.NewFromInt32(data.GetLongitude()).Div(decimal.NewFromInt32(10000000)).StringFixed(7),
		GasResistance:  decimal.NewFromInt32(int32(data.GetGasResistance())).Div(decimal.NewFromInt32(100)).StringFixed(2),
		Temperature:    decimal.NewFromInt32(data.GetTemperature()).Div(decimal.NewFromInt32(100)).StringFixed(2),
		Temperature2:   decimal.NewFromInt32(int32(data.GetTemperature2())).Div(decimal.NewFromInt32(100)).StringFixed(2),
		Pressure:       decimal.NewFromInt32(int32(data.GetPressure())).Div(decimal.NewFromInt32(100)).StringFixed(2),
		Humidity:       decimal.NewFromInt32(int32(data.GetHumidity())).Div(decimal.NewFromInt32(100)).StringFixed(2),
		Light:          decimal.NewFromInt32(int32(data.GetLight())).Div(decimal.NewFromInt32(100)).StringFixed(2),
		Gyroscope:      string(gyroscope),
		Accelerometer:  string(accelerometer),
		OperationTimes: db.NewOperationTimes(),
	}
	err = s.db.CreateDeviceRecord(dr)
	return errors.Wrapf(err, "failed to create senser data: %s", id)
}

func Run(db *db.DB, address string, client *ethclient.Client, ioidAddr, ioidRegistryAddr common.Address) error {
	ioidInstance, err := ioid.NewIoid(ioidAddr, client)
	if err != nil {
		return errors.Wrap(err, "failed to new ioid contract instance")
	}
	ioidRegistryInstance, err := ioidregistry.NewIoidregistry(ioidRegistryAddr, client)
	if err != nil {
		return errors.Wrap(err, "failed to new ioid registry contract instance")
	}
	s := &httpServer{
		engine:               gin.Default(),
		db:                   db,
		ioidInstance:         ioidInstance,
		ioidRegistryInstance: ioidRegistryInstance,
	}

	s.engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	s.engine.GET("/device", s.query)
	s.engine.POST("/device", s.receive)

	err = s.engine.Run(address)
	return errors.Wrap(err, "failed to start http server")
}
