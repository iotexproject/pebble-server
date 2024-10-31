package api

import (
	"context"
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/machinefi/ioconnect-go/pkg/ioconnect"
	"github.com/pkg/errors"

	"github.com/iotexproject/pebble-server/cmd/server/apitypes"
	"github.com/iotexproject/pebble-server/cmd/server/clients"
	"github.com/iotexproject/pebble-server/models"
	"github.com/iotexproject/pebble-server/modules/event"
)

type httpServer struct {
	ctx     context.Context
	engine  *gin.Engine
	jwk     *ioconnect.JWK
	clients *clients.Manager
}

func Run(ctx context.Context, jwk *ioconnect.JWK, clientMgr *clients.Manager, address string) error {
	s := &httpServer{
		ctx:     ctx,
		engine:  gin.Default(),
		jwk:     jwk,
		clients: clientMgr,
	}

	s.engine.POST("/issue_vc", s.issueJWTCredential)
	s.engine.POST("/device/data", s.verifyToken, s.receiveDeviceData)
	s.engine.GET("/device/query", s.verifyToken, s.queryDeviceState)
	s.engine.GET("/didDoc", s.didDoc)

	err := s.engine.Run(address)
	return errors.Wrap(err, "failed to start http server")
}

// verifyToken make sure the client token is issued by sequencer
func (s *httpServer) verifyToken(c *gin.Context) {
	tok := c.GetHeader("Authorization")
	if tok == "" {
		tok = c.Query("authorization")
	}

	if tok == "" {
		slog.Error("empty authorization token")
		return
	}

	tok = strings.TrimSpace(strings.Replace(tok, "Bearer", " ", 1))

	clientID, err := s.jwk.VerifyToken(tok)
	if err != nil {
		slog.Error("failed to verify token", "error", err)
		c.JSON(http.StatusUnauthorized, apitypes.NewErrRsp(errors.Wrap(err, "invalid credential token")))
		return
	}
	client := s.clients.ClientByIoID(clientID)
	if client == nil {
		slog.Error("failed to get client by ioid", "client_id", clientID)
		c.JSON(http.StatusUnauthorized, apitypes.NewErrRsp(errors.New("invalid credential token")))
		return
	}

	ctx := clients.WithClientID(c.Request.Context(), client)
	c.Request = c.Request.WithContext(ctx)
	slog.Info("verify token succeed", "client_id", clientID)
}

func (s *httpServer) receiveDeviceData(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		slog.Error("failed to read request body", "error", err)
		c.JSON(http.StatusInternalServerError, apitypes.NewErrRsp(errors.Wrap(err, "failed to read request body")))
		return
	}
	defer c.Request.Body.Close()

	// decrypt did comm message
	client := clients.ClientIDFrom(c.Request.Context())
	if client != nil {
		slog.Info("decrypted payload", "payload", string(payload))
		payload, err = s.jwk.Decrypt(payload, client.DID())
		if err != nil {
			slog.Error("failed to decrypt didcomm cipher data", "error", err)
			c.JSON(http.StatusBadRequest, apitypes.NewErrRsp(errors.Wrap(err, "failed to decrypt didcomm cipher data")))
			return
		}
		slog.Info("encrypted payload", "payload", string(payload))
	}
	payload, err = base64.RawURLEncoding.DecodeString(string(payload))
	if err != nil {
		slog.Error("failed to decode base64 data", "error", err)
		c.JSON(http.StatusBadRequest, apitypes.NewErrRsp(errors.Wrap(err, "failed to decode base64 data")))
		return
	}

	e := &event.DeviceData{}
	if err := e.Unmarshal(payload); err != nil {
		slog.Error("failed to unmarshal device data", "error", err)
		c.JSON(http.StatusBadRequest, apitypes.NewErrRsp(errors.Wrap(err, "failed to unmarshal request body")))
		return
	}
	e.Imei = client.DID()
	if err := e.Handle(s.ctx); err != nil {
		slog.Error("failed to handle device data", "error", err)
		c.JSON(http.StatusInternalServerError, apitypes.NewErrRsp(errors.Wrap(err, "failed to receive device data")))
		return
	}
	c.Status(http.StatusOK)
}

func (s *httpServer) queryDeviceState(c *gin.Context) {
	client := clients.ClientIDFrom(c.Request.Context())

	dev := &models.Device{ID: client.DID()}
	if err := event.FetchByPrimary(s.ctx, dev); err != nil {
		slog.Error("failed to query device", "error", err)
		c.JSON(http.StatusInternalServerError, apitypes.NewErrRsp(err))
		return
	}
	if dev.Status == models.CREATED {
		slog.Error("device is not propsaled", "device_id", dev.ID)
		c.JSON(http.StatusBadRequest, apitypes.NewErrRsp(errors.Errorf("device %s is not propsaled", dev.ID)))
		return
	}
	var (
		firmware string
		uri      string
		version  string
	)
	if parts := strings.Split(dev.RealFirmware, " "); len(parts) == 2 {
		app := &models.App{ID: parts[0]}
		err := event.FetchByPrimary(s.ctx, app)
		if err == nil {
			firmware = app.ID
			uri = app.Uri
			version = app.Version
		}
	}

	// meta := contexts.AppMeta().MustFrom(ctx)
	//pubType := "pub_DeviceQueryRsp"
	pubData := &struct {
		Status     int32  `json:"status"`
		Proposer   string `json:"proposer"`
		Firmware   string `json:"firmware,omitempty"`
		URI        string `json:"uri,omitempty"`
		Version    string `json:"version,omitempty"`
		ServerMeta string `json:"server_meta,omitempty"`
	}{
		Status:   dev.Status,
		Proposer: dev.Proposer,
		Firmware: firmware,
		URI:      uri,
		Version:  version,
		// ServerMeta: meta.String(),
	}

	// if client != nil {
	// 	slog.Info("encrypt response task query", "response", response)
	// 	cipher, err := s.jwk.EncryptJSON(response, client.KeyAgreementKID())
	// 	if err != nil {
	// 		c.JSON(http.StatusInternalServerError, apitypes.NewErrRsp(errors.Wrap(err, "failed to encrypt response when query task")))
	// 		return
	// 	}
	// 	c.Data(http.StatusOK, "application/octet-stream", cipher)
	// 	return
	// }

	c.JSON(http.StatusOK, pubData)
}

func (s *httpServer) didDoc(c *gin.Context) {
	if s.jwk == nil {
		c.JSON(http.StatusNotAcceptable, apitypes.NewErrRsp(errors.New("jwk is not config")))
		return
	}
	c.JSON(http.StatusOK, s.jwk.Doc())
}

func (s *httpServer) issueJWTCredential(c *gin.Context) {
	req := new(apitypes.IssueTokenReq)
	if err := c.ShouldBindJSON(req); err != nil {
		slog.Error("failed to bind request", "error", err)
		c.JSON(http.StatusBadRequest, apitypes.NewErrRsp(err))
		return
	}

	client := s.clients.ClientByIoID(req.ClientID)
	if client == nil {
		slog.Error("failed to get client", "client_id", req.ClientID)
		c.String(http.StatusForbidden, errors.Errorf("client is not register to ioRegistry").Error())
		return
	}

	token, err := s.jwk.SignToken(req.ClientID)
	if err != nil {
		slog.Error("failed to sign token", "client_id", req.ClientID, "error", err)
		c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to sign token").Error())
		return
	}
	slog.Info("token signed", "token", token)

	cipher, err := s.jwk.Encrypt([]byte(token), client.KeyAgreementKID())
	if err != nil {
		slog.Error("failed to encrypt", "client_id", req.ClientID, "error", err)
		c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to encrypt").Error())
		return
	}

	dev := &models.Device{
		ID:             client.DID(),
		Owner:          client.Owner().String(),
		Address:        common.HexToAddress(strings.TrimPrefix(client.DID(), "did:io:")).String(),
		Status:         models.CONFIRM,
		Proposer:       client.Owner().String(),
		OperationTimes: models.NewOperationTimes(),
	}
	if _, err := event.UpsertOnConflict(s.ctx, dev, "id", "owner", "proposer", "status", "updated_at"); err != nil {
		slog.Error("failed to upsert device", "client_id", req.ClientID, "error", err)
		c.String(http.StatusInternalServerError, errors.Wrap(err, "failed to upsert device").Error())
		return
	}

	c.Data(http.StatusOK, "application/json", cipher)
}
