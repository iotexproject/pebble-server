package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	ProjectID      uint64         `gorm:"index:task_fetch,not null"`
	InternalTaskID string         `gorm:"index:internal_task_id,not null"`
	MessageIDs     datatypes.JSON `gorm:"not null"`
	Signature      string         `gorm:"not null,default:''"`
}
