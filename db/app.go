package db

import (
	"bytes"
	"encoding/json"
	"log/slog"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"gorm.io/gorm/clause"
)

type App struct {
	ID      string `gorm:"primary_key"`
	Version string `gorm:"not null;default:''"`
	Uri     string `gorm:"not null;default:''"`
	Avatar  string `gorm:"not null;default:''"`
	Content string `gorm:"not null;default:''"`

	OperationTimes
}

func (*App) TableName() string { return "app" }

type firmwareData struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	URL     string `json:"url"`
}

var pebbleFirmwareKey = crypto.Keccak256Hash([]byte("pebble_firmware"))

func (d *DB) UpsertApp(projectID uint64, name string, key [32]byte, value []byte) error {
	if d.ioidProjectID != projectID {
		slog.Debug("not ioid project metadata", "project_id", projectID, "ioid_project_id", d.ioidProjectID)
		return nil
	}
	if !bytes.Equal(key[:], pebbleFirmwareKey.Bytes()) {
		slog.Error("failed to match pebble firmware key")
		return nil
	}

	firmware := &firmwareData{}
	if err := json.Unmarshal(value, firmware); err != nil {
		slog.Error("failed to unmarshal firmware data", "data", string(value), "error", err)
		return nil
	}

	t := App{
		ID:             name,
		Version:        firmware.Version,
		Uri:            firmware.URL,
		OperationTimes: NewOperationTimes(),
	}
	err := d.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"version", "uri", "updated_at"}),
	}).Create(&t).Error
	return errors.Wrap(err, "failed to upsert app")
}
