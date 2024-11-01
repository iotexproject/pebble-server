package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"

	"github.com/iotexproject/pebble-server/api"
	"github.com/iotexproject/pebble-server/cmd/server/config"
	"github.com/iotexproject/pebble-server/db"
)

func main() {
	cfg, err := config.Get()
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to get config"))
	}
	cfg.Print()
	slog.Info("pebble server config loaded")

	db, err := db.New(cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to new db"))
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
