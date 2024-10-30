package blockchain

import "github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/logger"

var l = logger.Default

func SetLogger(logger *logger.Logger) {
	l = logger
}
