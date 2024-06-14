package database

import (
	"net/url"

	"github.com/xoctopus/datatypex"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Postgres struct {
	Endpoint datatypex.Endpoint

	*gorm.DB `env:"-"`
}

func (d *Postgres) SetDefault() {
	if d.Endpoint.IsZero() {
		d.Endpoint = datatypex.Endpoint{
			Scheme:   "postgres",
			Host:     "127.0.0.1",
			Port:     5432,
			Base:     "pebble",
			Path:     "",
			Username: "",
			Password: "",
			Param:    url.Values{"sslmode": []string{"disable"}},
		}
	}
}

func (d *Postgres) Init() error {
	db, err := gorm.Open(postgres.Open(d.Endpoint.String()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}
	d.DB = db
	return nil
}
