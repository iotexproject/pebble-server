package models

type Account struct {
	ID     string `gorm:"primaryKey"`
	Name   string `gorm:"name"`
	Avatar string `gorm:"avatar"`
}
