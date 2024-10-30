package models

import "gorm.io/gorm"

type Message struct {
	gorm.Model
	MessageID      string `gorm:"index:message_id,not null"`
	ClientID       string `gorm:"index:message_fetch,not null,default:''"`
	ProjectID      uint64 `gorm:"index:message_fetch,not null"`
	ProjectVersion string `gorm:"index:message_fetch,not null,default:'0.0'"`
	Data           []byte `gorm:"size:4096"`
	InternalTaskID string `gorm:"index:internal_task_id,not null,default:''"`
}
