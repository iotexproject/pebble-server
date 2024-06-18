package models

import "time"

type AppV2 struct {
	ID         string    `gorm:"primary_key"`
	Slug       string    `gorm:"not null;default:''"`
	Logo       string    `gorm:"not null;default:''"`
	Author     string    `gorm:"not null;default:''"`
	Status     string    `gorm:"not null;default:''"`
	Content    string    `gorm:"not null;default:''"`
	Data       string    `gorm:"not null;default:'{}'"`
	Previews   string    `gorm:"not null;default:'[]'"`
	Date       time.Time `gorm:"not null;default:'';type:date"`
	CreatedAt  string    `gorm:"not null;default:''"`
	UpdatedAt  string    `gorm:"not null;default:''"`
	URI        string    `gorm:"not null;default:''"`
	Category   int32     `gorm:"not null;default:0"`
	DirectLink string    `gorm:"not null;default:''"`
	Order      int32     `gorm:"not null;default:0"`
	Firmware   string    `gorm:"not null;default:''"`

	OperationTimes
}

func (*AppV2) TableName() string { return "app_v2" }
