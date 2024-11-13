package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"

	"github.com/iotexproject/pebble-server/api"
	"github.com/iotexproject/pebble-server/cmd/server/config"
	"github.com/iotexproject/pebble-server/db"
	"github.com/iotexproject/pebble-server/monitor"
)

func main() {
	cfg, err := config.Get()
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to get config"))
	}
	cfg.Print()
	slog.Info("pebble server config loaded")

	prv, err := crypto.HexToECDSA(cfg.PrvKey)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to parse private key"))
	}

	db, err := db.New(cfg.DatabaseDSN, cfg.IoIDProjectID)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to new db"))
	}

	client, err := ethclient.Dial(cfg.ChainEndpoint)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to dial chain endpoint"))
	}

	if err := monitor.Run(
		&monitor.Handler{
			ScannedBlockNumber:       db.ScannedBlockNumber,
			UpsertScannedBlockNumber: db.UpsertScannedBlockNumber,
			UpsertProjectMetadata:    db.UpsertApp,
		},
		common.HexToAddress(cfg.ProjectContractAddr),
		cfg.BeginningBlockNumber,
		client,
	); err != nil {
		log.Fatal(errors.Wrap(err, "failed to run contract monitor"))
	}

	go func() {
		if err := api.Run(db, cfg.ServiceEndpoint, client, prv, common.HexToAddress(cfg.IoIDContractAddr), common.HexToAddress(cfg.IoIDRegistryContractAddr)); err != nil {
			log.Fatal(err)
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
}
