package blockchain

import "github.com/iotexproject/pebble-server/middlewares/logger"

var l = logger.Default

func SetLogger(logger *logger.Logger) {
	l = logger
}
