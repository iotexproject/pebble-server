package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/xoctopus/confx/confapp"
	"github.com/xoctopus/confx/confmws/confmqtt"
	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/misc/must"

	"github.com/machinefi/sprout-pebble-sequencer/cmd/sequencer/commands"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/alert"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/blockchain"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/crypto"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/database"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/logger"
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
		DryRun         bool
		MqttBroker     *confmqtt.Broker
		Database       *database.Postgres
		Blockchain     *blockchain.Blockchain
		Logger         *logger.Logger
		PrivateKey     *crypto.EcdsaPrivateKey
		ProjectID      uint64
		ProjectVersion string
		WhiteList      contexts.WhiteList
		LarkAlert      *alert.LarkAlert
	}{
		DryRun:     false,
		Logger:     &logger.Logger{Level: slog.LevelDebug},
		Blockchain: &blockchain.Blockchain{Contracts: contracts},
		MqttBroker: &confmqtt.Broker{},
		Database:   &database.Postgres{},
		// from sprout default sequencer, to make coordinator validate sequencer signature
		PrivateKey: &crypto.EcdsaPrivateKey{
			Hex: "dbfe03b0406549232b8dccc04be8224fcc0afa300a33d4f335dcfdfead861c85",
		},
		LarkAlert: &alert.LarkAlert{
			Env:     "PROD",
			Project: Name,
			Version: Version,
		},
		// WhiteList: contexts.WhiteList{"103381234567407"},
	}
	ctx context.Context
)

func init() {
	meta := confapp.Meta{
		Name:     Name,
		Feature:  Feature,
		Version:  Version,
		CommitID: CommitID,
		Date:     Date,
	}
	app = confapp.NewAppContext(
		confapp.WithBuildMeta(meta),
		confapp.WithMainRoot("."),
		confapp.WithMainExecutor(Main),
	)

	app.Conf(config)
	must.BeTrueWrap(
		config.ProjectVersion != "" && config.ProjectID != 0,
		"project id and version is required",
	)

	ctx = contextx.WithContextCompose(
		contexts.Logger().Compose(config.Logger),
		contexts.Blockchain().Compose(config.Blockchain),
		contexts.Database().Compose(config.Database),
		contexts.MqttBroker().Compose(config.MqttBroker),
		contexts.ProjectID().Compose(config.ProjectID),
		contexts.ProjectVersion().Compose(config.ProjectVersion),
		contexts.PrivateKey().Compose(config.PrivateKey),
		contexts.IMEIFilter().Compose(config.WhiteList),
		contexts.LarkAlert().Compose(config.LarkAlert),
		contexts.AppMeta().Compose(&meta),
	)(context.Background())

	app.AddCommand(commands.Migrate(ctx))
	app.AddCommand(commands.GenerateSproutConfig(ctx))
}

// Main app main entry
func Main() error {
	_ = config.LarkAlert.Push("service started", "")

	blockchain.SetLogger(config.Logger)
	if err := config.Blockchain.RunMonitor(); err != nil {
		config.Logger.Error(err, "failed to start tx monitor")
		return err
	}
	event.InitRunner(ctx)()
	defer config.Blockchain.Close()

	go RunDebugServer(ctx)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	_ = <-sig

	config.LarkAlert.Push("service stopped", "")
	return nil
}

func main() {
	if err := app.Command.Execute(); err != nil {
		app.PrintErrln(err)
	}
	config.Blockchain.Close()
	os.Exit(-1)
}
