package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xoctopus/x/misc/must"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
)

// RunDebugServer enable simple http server for debugging
func RunDebugServer(ctx context.Context, addr string) {
	eng := gin.Default()
	eng.Handle(
		http.MethodGet, "/debug/monitor-info",
		func(c *gin.Context) {
			bc := must.BeTrueV(contexts.BlockchainFromContext(ctx))
			monitors := bc.MonitorsInfo()
			c.JSON(http.StatusOK, monitors)
		},
	)
	eng.Run(addr)
}
