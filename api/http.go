package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/iotexproject/pebble-server/db"
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

type httpServer struct {
	engine *gin.Engine
	db     *db.DB
}

func (s *httpServer) query(c *gin.Context) {
	req := &queryReq{}
	if err := c.ShouldBindJSON(req); err != nil {
		slog.Error("failed to bind request", "error", err)
		c.JSON(http.StatusBadRequest, newErrResp(errors.Wrap(err, "invalid request payload")))
		return
	}

	sigStr := req.Signature
	req.Signature = ""

	reqJson, err := json.Marshal(req)
	if err != nil {
		slog.Error("failed to marshal request into json format", "error", err)
		c.JSON(http.StatusInternalServerError, newErrResp(errors.Wrap(err, "failed to process request data")))
		return
	}

	sig, err := hexutil.Decode(sigStr)
	if err != nil {
		slog.Error("failed to decode signature from hex format", "signature", sigStr, "error", err)
		c.JSON(http.StatusBadRequest, newErrResp(errors.Wrap(err, "invalid signature format")))
		return
	}

	h := crypto.Keccak256Hash(reqJson)
	sigpk, err := crypto.SigToPub(h.Bytes(), sig)
	if err != nil {
		slog.Error("failed to recover public key from signature", "error", err)
		c.JSON(http.StatusBadRequest, newErrResp(errors.Wrap(err, "invalid signature; could not recover public key")))
		return
	}

	owner := crypto.PubkeyToAddress(*sigpk)

	d, err := s.db.Device(req.DeviceID)
	if err != nil {
		slog.Error("failed to query device", "error", err, "device_id", req.DeviceID)
		c.JSON(http.StatusInternalServerError, newErrResp(errors.Wrap(err, "failed to query device")))
		return
	}
	if d == nil {
		slog.Error("device not exist", "device_id", req.DeviceID)
		c.JSON(http.StatusBadRequest, newErrResp(errors.New("device not exist")))
		return
	}
	if d.Owner != owner.String() {
		slog.Error("no permission to access the device", "device_id", req.DeviceID)
		c.JSON(http.StatusForbidden, newErrResp(errors.New("no permission to access the device")))
		return
	}

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

}

func Run(db *db.DB, address string) error {
	s := &httpServer{
		engine: gin.Default(),
		db:     db,
	}

	s.engine.GET("/device", s.query)
	s.engine.POST("/device", s.receive)

	err := s.engine.Run(address)
	return errors.Wrap(err, "failed to start http server")
}
