package db

type Bank struct {
	Address string `gorm:"primary_key"`
	Balance string `gorm:"not null;default:'0'"`

	OperationTimes
}

func (*Bank) TableName() string { return "bank" }
