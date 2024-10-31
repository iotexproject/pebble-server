package db

type App struct {
	ID      string `gorm:"primary_key"`
	Version string `gorm:"not null;default:''"`
	Uri     string `gorm:"not null;default:''"`
	Avatar  string `gorm:"not null;default:''"`
	Content string `gorm:"not null;default:''"`

	OperationTimes
}

func (*App) TableName() string { return "app" }
