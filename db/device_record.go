package db

import (
	"fmt"

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
	sql := `SELECT device_record_id FROM device_record_geo_locations 
              WHERE ST_DWithin(
              geom,
              ST_MakePoint(%s, %s)::geography,
              5000
            );`

	ids := []string{}
	if err := d.db.Raw(fmt.Sprintf(sql, longitude, latitude)).Scan(&ids).Error; err != nil {
		return nil, errors.Wrap(err, "failed to query device record geo data")
	}
	if len(ids) != 0 {
		t := &DeviceRecord{}
		if err := d.db.Where("id IN ?", ids).Order("timestamp DESC").First(&t).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, nil
			}
			return nil, errors.Wrap(err, "failed to query device record")
		}
		return t, nil
	}
	oldIDs := []string{}
	if err := d.oldDB.Raw(fmt.Sprintf(sql, longitude, latitude)).Scan(&ids).Error; err != nil {
		return nil, errors.Wrap(err, "failed to query device record geo data from old db")
	}
	if len(oldIDs) == 0 {
		return nil, nil
	}
	t := &DeviceRecord{}
	if err := d.oldDB.Where("id IN ?", oldIDs).Order("timestamp DESC").First(&t).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to query device record from old db")
	}
	return t, nil
}

func (d *DB) CreateDeviceRecord(t *DeviceRecord) error {
	err := d.db.Create(t).Error
	return errors.Wrap(err, "failed to create device record")
}
