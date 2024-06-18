package models

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
	RealFirmware           string `gorm:"not null;default:0"`
	State                  int32  `gorm:"not null;default:0"`
	Type                   int32  `gorm:"not null;default:0"`
	Configurable           bool   `gorm:"not null;default:0;default:true"`

	OperationTimes
}

func (*Device) TableName() string { return "device" }
