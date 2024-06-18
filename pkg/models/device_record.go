package models

type DeviceRecord struct {
	ID            string `gorm:"primary_key"`
	Imei          string `gorm:"not null"`
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
	Timestamp     int64  `gorm:"not null;default:0"`

	OperationTimes
}

func (*DeviceRecord) TableName() string { return "device_record" }
