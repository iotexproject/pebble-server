package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/xoctopus/confx/confapp"
	"github.com/xoctopus/confx/confmws/confmqtt"
	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/misc/must"

	"github.com/machinefi/ioconnect-go/pkg/ioconnect"
	"github.com/machinefi/sprout-pebble-sequencer/cmd/sequencer/api"
	"github.com/machinefi/sprout-pebble-sequencer/cmd/sequencer/clients"
	"github.com/machinefi/sprout-pebble-sequencer/cmd/sequencer/commands"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/alert"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/blockchain"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/crypto"
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
		DryRun                          bool
		MqttBroker                      *confmqtt.Broker
		Database                        *database.Postgres
		Blockchain                      *blockchain.Blockchain
		Logger                          *logger.Logger
		PrivateKey                      *crypto.EcdsaPrivateKey
		ProjectID                       uint64
		ProjectVersion                  string
		WhiteList                       contexts.WhiteList
		LarkAlert                       *alert.LarkAlert
		MqttClientID                    string
		JwkSecret                       string
		IoIDRegistryEndpoint            string
		IoIDRegistryContractAddress     string
		ProjectClientContractAddress    string
		W3bstreamProjectContractAddress string
		ChainEndpoint                   string
		IoIDProjectID                   uint64
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
		MqttClientID:                    uuid.NewString(),
		JwkSecret:                       "R3QNJihYLjtcaxALSTsKe1cYSX0pS28wZitFVXE4Y2klf2hxVCczYHw2dVg4fXJdSgdCcnM4PgV1aTo9DwYqEw==",
		IoIDRegistryEndpoint:            "did.iotex.me",
		IoIDRegistryContractAddress:     "0x0A7e595C7889dF3652A19aF52C18377bF17e027D",
		ProjectClientContractAddress:    "0xF4d6282C5dDD474663eF9e70c927c0d4926d1CEb",
		W3bstreamProjectContractAddress: "0x6AfCB0EB71B7246A68Bb9c0bFbe5cD7c11c4839f",
		ChainEndpoint:                   "https://babel-api.testnet.iotex.io",
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
		config.ProjectVersion != "" && config.ProjectID != 0 && config.IoIDProjectID != 0,
		"project id, version and ioID project id is required",
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
		contexts.DryRun().Compose(config.DryRun),
		contexts.MqttClientID().Compose(config.MqttClientID),
		contexts.IoIDProjectID().Compose(config.IoIDProjectID),
	)(context.Background())

	app.AddCommand(commands.Migrate(ctx))
	app.AddCommand(commands.GenerateSproutConfig(ctx))
}

func runHTTP(ctx context.Context) {
	secrets := ioconnect.JWKSecrets{}
	if err := secrets.UnmarshalText([]byte(config.JwkSecret)); err != nil {
		panic(errors.Wrap(err, "invalid jwk secrets from flag"))
	}
	jwk, err := ioconnect.NewJWKBySecret(secrets)
	if err != nil {
		panic(errors.Wrap(err, "failed to new jwk from secrets"))
	}
	clientMgr, err := clients.NewManager(config.ProjectClientContractAddress, config.IoIDRegistryContractAddress, config.W3bstreamProjectContractAddress, config.IoIDRegistryEndpoint, config.ChainEndpoint)
	if err != nil {
		panic(errors.Wrap(err, "failed to new clients manager"))
	}
	go func() {
		if err := api.NewHttpServer(ctx, jwk, clientMgr).Run(":9000"); err != nil {
			panic(err)
		}
	}()
}

// Main app main entry
func Main() error {
	_ = config.LarkAlert.Push("service started", "")

	db := contexts.Database().MustFrom(ctx)
	if err := db.AutoMigrate(
		&models.Account{},
		&models.App{},
		&models.AppV2{},
		&models.Bank{},
		&models.BankRecord{},
		&models.Device{},
		&models.DeviceRecord{},
		&models.Task{},
		&models.Message{},
	); err != nil {
		slog.Error("failed to migrate database", "error", err)
		return err
	}

	blockchain.SetLogger(config.Logger)
	if err := config.Blockchain.RunMonitor(); err != nil {
		config.Logger.Error(err, "failed to start tx monitor")
		return err
	}
	event.InitRunner(ctx)()
	defer config.Blockchain.Close()

	go RunDebugServer(ctx)
	runHTTP(ctx)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

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
