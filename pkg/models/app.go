package models

type App struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Uri     string `json:"uri"`
	Avatar  string `json:"avatar"`
	Content string `json:"content"`

	OperationTimes
}

func (*App) TableName() string { return "app" }
