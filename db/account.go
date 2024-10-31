package db

type Account struct {
	ID     string `gorm:"primary_key"`
	Name   string `gorm:"not null"`
	Avatar string `gorm:"not null"`

	OperationTimes
}

func (*Account) TableName() string { return "account" }
