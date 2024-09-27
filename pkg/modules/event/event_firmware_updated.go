package event

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
)

func init() {
	f := func() Event { return &FirmwareUpdated{} }
	e := f()
	registry(e.Topic(), f)
}

type FirmwareUpdated struct {
	Name    string
	Version string
	Uri     string
	Avatar  string
}

func (e *FirmwareUpdated) Source() enums.EventSourceType {
	return enums.EVENT_SOURCE_TYPE__BLOCKCHAIN
}

func (e *FirmwareUpdated) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *FirmwareUpdated) ContractID() string { return enums.CONTRACT__PEBBLE_FIRMWARE }

func (e *FirmwareUpdated) EventName() string { return "AddMetadata" }

type (
	AddMetadataEvent struct {
		ProjectId *big.Int
		Name      string
		Key       [32]byte
		Value     []byte
	}

	FirmwareData struct {
		Name    string `json:"name"`
		Version int    `json:"version"`
		URL     string `json:"url"`
	}
)

func (e *FirmwareUpdated) Unmarshal(v any) error {
	ame := &AddMetadataEvent{}

	if err := v.(TxEventUnmarshaler).UnmarshalTx(e.EventName(), ame); err != nil {
		return err
	}

	pebbleKey := crypto.Keccak256Hash([]byte("pebble_firmware"))
	if !bytes.Equal(ame.Key[:], pebbleKey[:]) {
		return errors.Errorf("key mismatch: %x != %x", ame.Key, pebbleKey)
	}

	var firmware FirmwareData

	err := json.Unmarshal(ame.Value, &firmware)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal firmware data: %s", string(ame.Value))
	}

	e.Name = firmware.Name
	e.Version = strconv.Itoa(firmware.Version)
	e.Uri = firmware.URL

	return nil
}

func (e *FirmwareUpdated) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	app := &models.App{
		ID:             e.Name,
		Version:        e.Version,
		Uri:            e.Uri,
		Avatar:         e.Avatar,
		OperationTimes: models.NewOperationTimes(),
	}
	_, err = UpsertOnConflict(ctx, app, "id", "version", "uri", "avatar", "updated_at")
	if err != nil {
		return errors.Wrapf(err, "failed to upsert app: %s", app.ID)
	}

	// meta := contexts.AppMeta().MustFrom(ctx)
	pubType := "pub_FirmwareUpdatedRsp"
	pubData := &struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Uri     string `json:"uri"`
		Avatar  string `json:"avatar"`
		// ServerMeta string `json:"meta"`
	}{
		Name:    app.ID,
		Version: app.Version,
		Uri:     app.Uri,
		Avatar:  app.Avatar,
		// ServerMeta: meta.String(),
	}
	return errors.Wrapf(
		PublicMqttMessage(ctx, pubType, "device/app_update/"+app.ID, pubData),
		"failed to publish %s", pubType,
	)
}
