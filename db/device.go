package db

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	CREATED int32 = iota
	PROPOSAL
	CONFIRM
)

type Device struct {
	ID                     string `gorm:"primary_key"`
	Name                   string `gorm:"not null;default:''"`
	Owner                  string `gorm:"not null;default:''"`
	Address                string `gorm:"not null;default:''"`
	Avatar                 string `gorm:"not null;default:''"`
	Status                 int32  `gorm:"not null;default:0"`
	Proposer               string `gorm:"not null;default:''"`
	Firmware               string `gorm:"not null;default:''"`
	Config                 string `gorm:"not null;default:''"`
	TotalGas               int32  `gorm:"not null;default:0"`
	BulkUpload             int32  `gorm:"not null;default:0"`
	DataChannel            int32  `gorm:"not null;default:0"`
	UploadPeriod           int32  `gorm:"not null;default:0"`
	BulkUploadSamplingCnt  int32  `gorm:"not null;default:0"`
	BulkUploadSamplingFreq int32  `gorm:"not null;default:0"`
	Beep                   int32  `gorm:"not null;default:0"`
	RealFirmware           string `gorm:"not null;default:''"`
	State                  int32  `gorm:"not null;default:0"`
	Type                   int32  `gorm:"not null;default:0"`
	Configurable           bool   `gorm:"not null;default:0;default:true"`

	OperationTimes
}

func (*Device) TableName() string { return "device" }

func (d *DB) Device(id string) (*Device, error) {
	t := Device{}
	if err := d.db.Where("id = ?", id).First(&t).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to query device")
	}
	return &t, nil
}

func (d *DB) UpsertDevice(t *Device) error {
	err := d.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"owner", "address", "status", "proposer", "updated_at"}),
	}).Create(t).Error
	return errors.Wrap(err, "failed to upsert device")
}

func (d *DB) UpdateByID(id string, values map[string]any) error {
	err := d.db.Model(&Device{}).Where("id = ?", id).Updates(values).Error
	return errors.Wrap(err, "failed to update device")
}
