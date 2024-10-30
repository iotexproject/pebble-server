package commands

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/xoctopus/x/misc/must"

	"github.com/iotexproject/pebble-server/contexts"
	"github.com/iotexproject/pebble-server/models"
)

func Migrate(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "migrate database",
		Run: func(cmd *cobra.Command, args []string) {
			db := contexts.Database().MustFrom(ctx)
			must.NoErrorWrap(db.AutoMigrate(
				&models.Account{},
				&models.App{},
				&models.AppV2{},
				&models.Bank{},
				&models.BankRecord{},
				&models.Device{},
				&models.DeviceRecord{},
				&models.Task{},
				&models.Message{},
			), "failed to migrate database")
		},
	}
}
