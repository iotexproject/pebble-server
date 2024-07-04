package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/xoctopus/x/misc/must"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
)

// RunDebugServer enable simple http server for debugging
func RunDebugServer(ctx context.Context) {
	// addr := contexts.ServerAddrFromContext(ctx)
	eng := gin.Default()
	eng.Handle(
		http.MethodGet, "/debug/monitor",
		func(c *gin.Context) {
			bc := must.BeTrueV(contexts.BlockchainFromContext(ctx))
			meta := bc.MonitorMeta()
			c.JSON(http.StatusOK, meta)
		},
	)
	eng.Handle(
		http.MethodGet, "/env/:key",
		func(c *gin.Context) {
			key := c.Param("key")
			val := os.Getenv(key)
			if val == "" {
				c.Status(http.StatusNotFound)
				return
			}
			c.String(http.StatusOK, val)
			return
		},
	)
	eng.Handle(
		http.MethodGet, "/version",
		func(c *gin.Context) {
			c.JSON(http.StatusOK, map[string]string{
				"service_name": Name,
				"feature":      Feature,
				"version":      Version,
				"commit_id":    CommitID,
				"build_at":     Date,
			})
		},
	)
	eng.Run(":80")
}
