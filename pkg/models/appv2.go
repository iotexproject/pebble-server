package models

type AppV2 struct {
	ID         string `json:"id"`
	Slug       string `json:"slug"`
	Logo       string `json:"logo"`
	Author     string `json:"author"`
	Status     string `json:"status"`
	Content    string `json:"content"`
	Data       string `json:"data"`
	Previews   string `json:"previews"`
	Date       string `json:"date"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	URI        string `json:"uri"`
	Category   int32  `json:"category"`
	DirectLink string `json:"direct_link"`
	Order      int32  `json:"order"`
	Firmware   string `json:"firmware"`

	OperationTimes
}

func (*AppV2) TableName() string { return "app_v2" }
