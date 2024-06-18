package models

type DeviceRecord struct {
	ID            string `json:"id"`             // id varchar(64)
	Imei          string `json:"imei"`           // imei varchar(64)
	Operator      string `json:"operator"`       // operator varchar(64)
	Snr           string `json:"snr"`            // snr varchar(12)
	Vbat          string `json:"vbat"`           // vbat varchar(12)
	GasResistance string `json:"gas_resistance"` // gas_resistance varchar(12)
	Temperature   string `json:"temperature"`    // temperature varchar(12)
	Temperature2  string `json:"temperature2"`   // temperature2 varchar(12)
	Pressure      string `json:"pressure"`       // pressure varchar(12)
	Humidity      string `json:"humidity"`       // humidity varchar(12)
	Light         string `json:"light"`          // light varchar(12)
	Gyroscope     string `json:"gyroscope"`      // gyroscope varchar(128)
	Accelerometer string `json:"accelerometer"`  // accelerometer varchar(128)
	Latitude      string `json:"latitude"`       // latitude varchar(32)
	Longitude     string `json:"longitude"`      // longitude varchar(32)
	Signature     string `json:"signature"`      // signature varchar(256)
	Timestamp     int64  `json:"timestamp"`      // timestamp int(11)

	OperationTimes
}

func (*DeviceRecord) TableName() string { return "device_record" }
