package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/xoctopus/confx/confapp"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/alert"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/database"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/logger"
)

var (
	Name     = "replication-monitor"
	Feature  string
	Version  string
	CommitID string
	Date     string

	app    *confapp.AppCtx
	config = &struct {
		Database  *database.Postgres
		Logger    *logger.Logger
		LarkAlert *alert.LarkAlert
	}{
		Logger:    &logger.Logger{Level: slog.LevelDebug},
		Database:  &database.Postgres{},
		LarkAlert: &alert.LarkAlert{},
	}
)

func init() {
	app = confapp.NewAppContext(
		confapp.WithBuildMeta(confapp.Meta{
			Name:     Name,
			Feature:  Feature,
			Version:  Version,
			CommitID: CommitID,
			Date:     Date,
		}),
		confapp.WithMainRoot("."),
		confapp.WithDefaultConfigGenerator(),
		confapp.WithMainExecutor(Main),
	)

	app.Conf(config)
}

type SubscriptionStat struct {
	SubID           string `gorm:"column:subid"`
	SubName         string `gorm:"column:subname"`
	ApplyErrorCount int64  `gorm:"column:apply_error_count"`
	SyncErrorCount  int64  `gorm:"column:sync_error_count"`
}

func (SubscriptionStat) TableName() string {
	return "pg_stat_subscription_stats"
}

func Main() error {
	d := config.Database
	l := config.Logger

	for {
		stats := make([]*SubscriptionStat, 0)
		msg := ""
		err := d.Find(&stats).Error
		if err != nil {
			l.Error(err, "failed to query subscription stat")
			goto TryLater
		}
		for _, s := range stats {
			if s.ApplyErrorCount > 0 {
				msg += fmt.Sprintf("%s: apply_error_count: %d\n", s.SubName, s.ApplyErrorCount)
				l.Info("subscription error", "name", s.SubName, "apply_error_count", s.ApplyErrorCount)
			}
			if s.SyncErrorCount > 0 {
				msg += fmt.Sprintf("%s: apply_error_count: %d\n", s.SubName, s.SyncErrorCount)
				l.Info("subscription error", "name", s.SubName, "sync_error_count", s.SyncErrorCount)
			}
			l.Info("subscription", "name", s.SubName, "stat", "ok")
		}
		if len(msg) > 0 {
			config.LarkAlert.Push("replication error", msg)
		}
	TryLater:
		time.Sleep(time.Minute)
		continue
	}
}

func main() {
	if err := app.Command.Execute(); err != nil {
		app.PrintErrln(err)
	}
	os.Exit(-1)
}
