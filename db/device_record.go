package db

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type DeviceRecord struct {
	ID            string `gorm:"primary_key"`
	Imei          string `gorm:"index:device_record_imei;not null"`
	Operator      string `gorm:"not null"`
	Snr           string `gorm:"not null;type:numeric(10,2);default:0"`
	Vbat          string `gorm:"not null;type:numeric(10,2);default:0"`
	GasResistance string `gorm:"not null;type:numeric(10,2);default:0"`
	Temperature   string `gorm:"not null;type:numeric(10,2);default:0"`
	Temperature2  string `gorm:"not null;type:numeric(10,2);default:0"`
	Pressure      string `gorm:"not null;type:numeric(10,2);default:0"`
	Humidity      string `gorm:"not null;type:numeric(10,2);default:0"`
	Light         string `gorm:"not null;type:numeric(10,2);default:0"`
	Gyroscope     string `gorm:"not null;default:''"`
	Accelerometer string `gorm:"not null;default:''"`
	Latitude      string `gorm:"not null;default:0"`
	Longitude     string `gorm:"not null;default:0"`
	Signature     string `gorm:"not null;default:''"`
	Timestamp     int64  `gorm:"index:device_record_timestamp;not null;default:0"`

	OperationTimes
}

func (*DeviceRecord) TableName() string { return "device_record" }

func (d *DB) QueryDeviceRecord(latitude, longitude string) (*DeviceRecord, error) {
	t := &DeviceRecord{}
	if err := d.db.Where("latitude = ?", latitude).Where("longitude = ?", longitude).Order("timestamp DESC").First(&t).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to query device record")
	}
	return t, nil
}

func (d *DB) CreateDeviceRecord(t *DeviceRecord) error {
	err := d.db.Create(t).Error
	return errors.Wrap(err, "failed to create device record")
}
