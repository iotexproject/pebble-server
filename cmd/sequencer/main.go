package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum/common"
	"github.com/xoctopus/confx/confapp"
	"github.com/xoctopus/confx/confmws/confmqtt"
	"github.com/xoctopus/x/contextx"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/blockchain"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/database"
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
		MqttBroker confmqtt.Broker
		Database   database.Postgres
		Blockchain blockchain.Blockchain
		Contracts  map[string]common.Address
	}{
		Contracts: map[string]common.Address{
			"account":  common.HexToAddress("0x189e2ED6EAfBCeAF938d049cf3685828b5493952"),
			"firmware": common.HexToAddress("0xC5F406c42C96e68756311Dad49dE99B0f4A1A722"),
			"pebble":   common.HexToAddress("0xC9D7D9f25b98119DF5b2303ac0Df6b15C982BbF5"),
			"bank":     common.HexToAddress("0xb86f97D494EEf8c6d618ee2049419eE0Ce843F28"),
		},
	}
	ctx context.Context
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
		confapp.WithPreRunner(
			event.InitRunner(ctx),
		),
	)

	app.Conf(config)

	ctx = contextx.WithContextCompose(
		contexts.WithBlockchainContext(&config.Blockchain),
		contexts.WithDatabaseContext(&config.Database),
		contexts.WithMqttBrokerContext(&config.MqttBroker),
		contexts.WithContractsContext(config.Contracts),
	)(context.Background())
}

func Main() error {
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)
	<-sig

	return nil
}

func main() {
	if err := app.Command.Execute(); err != nil {
		app.PrintErrln(err)
		os.Exit(-1)
	}
}
