package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/xoctopus/confx/confapp"
	"github.com/xoctopus/confx/confmws/confmqtt"
	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/misc/must"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/blockchain"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/database"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/logger"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/event"
)

var (
	Name     = "pebble-sequencer"
	Feature  string
	Version  string
	CommitID string
	Date     string

	app    *confapp.AppCtx
	config = &struct {
		MqttBroker *confmqtt.Broker
		Database   *database.Postgres
		Blockchain *blockchain.Blockchain
		Logger     *logger.Logger
		ServerPort uint16
	}{
		Logger:     &logger.Logger{Level: slog.LevelDebug},
		Blockchain: &blockchain.Blockchain{Contracts: contracts},
		MqttBroker: &confmqtt.Broker{},
		Database:   &database.Postgres{},
		ServerPort: 6666,
	}
	ctx context.Context
)

func init() {
	ctx = contextx.WithContextCompose(
		contexts.WithLoggerContext(config.Logger),
		contexts.WithBlockchainContext(config.Blockchain),
		contexts.WithDatabaseContext(config.Database),
		contexts.WithMqttBrokerContext(config.MqttBroker),
	)(context.Background())

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

	app.AddCommand(&cobra.Command{
		Use:   "migrate",
		Short: "migrate database",
		Run: func(cmd *cobra.Command, args []string) {
			Migrate(ctx)
		},
	})
}

func Main() error {
	if err := config.Blockchain.RunMonitors(); err != nil {
		config.Logger.Error(err, "failed to start tx monitor")
	}
	event.InitRunner(ctx)()
	go RunDebugServer(ctx, fmt.Sprintf(":%d", config.ServerPort))

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	<-sig

	return nil
}

func Migrate(ctx context.Context) {
	db := must.BeTrueV(contexts.DatabaseFromContext(ctx))
	must.NoErrorWrap(db.AutoMigrate(
		&models.Account{},
		&models.App{},
		&models.AppV2{},
		&models.Bank{},
		&models.BankRecord{},
		&models.Device{},
		&models.DeviceRecord{},
	), "failed to migrate database")
}

func main() {
	if err := app.Command.Execute(); err != nil {
		app.PrintErrln(err)
		os.Exit(-1)
	}
}
