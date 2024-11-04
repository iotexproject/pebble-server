package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/common"
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

	db, err := db.New(cfg.DatabaseDSN, cfg.IoIDProjectID)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to new db"))
	}

	if err := monitor.Run(
		&monitor.Handler{
			ScannedBlockNumber:       db.ScannedBlockNumber,
			UpsertScannedBlockNumber: db.UpsertScannedBlockNumber,
			UpsertProjectMetadata:    db.UpsertApp,
		},
		&monitor.ContractAddr{
			Project: common.HexToAddress(cfg.ProjectContractAddr),
		},
		cfg.BeginningBlockNumber,
		cfg.ChainEndpoint,
	); err != nil {
		log.Fatal(errors.Wrap(err, "failed to run contract monitor"))
	}

	go func() {
		if err := api.Run(db, cfg.ServiceEndpoint); err != nil {
			log.Fatal(err)
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
}
