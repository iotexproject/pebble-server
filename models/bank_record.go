package models

const (
	BankRecodeDeposit int32 = iota
	BankRecodeWithdraw
	BankRecodePaid
)

type BankRecord struct {
	ID        string `gorm:"primary_key"`
	From      string `gorm:"not null;default:''"`
	To        string `gorm:"not null;default:''"`
	Amount    string `gorm:"not null;default:''"`
	Timestamp int64  `gorm:"not null;default:0"`
	Type      int32  `gorm:"not null;default:0"`

	OperationTimes
}

func (*BankRecord) TableName() string { return "bank_record" }
