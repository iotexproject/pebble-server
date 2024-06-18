package models

type Account struct {
	ID     string
	Name   string
	Avatar string

	OperationTimes
}

func (*Account) TableName() string { return "account" }
