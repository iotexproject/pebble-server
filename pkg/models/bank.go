package models

type Bank struct {
	Address string `json:"address"`
	Balance string `json:"balance"`

	OperationTimes
}

func (*Bank) TableName() string { return "bank" }
