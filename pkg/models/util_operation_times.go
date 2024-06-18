package models

import "time"

func NewOperationTimes() OperationTimes {
	return OperationTimes{
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

type OperationTimes struct {
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}
