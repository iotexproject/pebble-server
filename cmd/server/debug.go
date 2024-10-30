package main

import (
	"context"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/xoctopus/datatypex"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
)

// RunDebugServer enable simple http server for debugging
func RunDebugServer(ctx context.Context) {
	// addr := contexts.ServerAddrFromContext(ctx)
	eng := gin.Default()
	eng.Handle(
		http.MethodGet, "/monitor",
		func(c *gin.Context) {
			bc := contexts.Blockchain().MustFrom(ctx)
			meta := bc.MonitorMeta()
			c.JSON(http.StatusOK, meta)
		},
	)
	eng.Handle(
		http.MethodGet, "/mqtt",
		func(c *gin.Context) {
			b := contexts.MqttBroker().MustFrom(ctx)
			c.JSON(http.StatusOK, b)
		},
	)
	eng.Handle(
		http.MethodGet, "/whitelist",
		func(c *gin.Context) {
			filter := contexts.IMEIFilter().MustFrom(ctx)
			c.JSON(http.StatusOK, filter)
		},
	)
	eng.Handle(
		http.MethodGet, "/project",
		func(c *gin.Context) {
			c.JSON(http.StatusOK, &struct {
				ID      uint64 `json:"id"`
				Version string `json:"version"`
			}{
				ID:      contexts.ProjectID().MustFrom(ctx),
				Version: contexts.ProjectVersion().MustFrom(ctx),
			})
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
		http.MethodGet, "/envs",
		func(c *gin.Context) {
			keys := make([]string, 0)
			vars := os.Environ()
			for _, kv := range vars {
				parts := strings.Split(kv, "=")
				if parts[0] == "PEBBLE_SEQUENCER__Database_Endpoint" {
					println(parts[0], parts[1])
				}
				if len(parts) >= 2 && strings.HasPrefix(parts[0], "PEBBLE_SEQUENCER__") {
					keys = append(keys, parts[0])
				}
			}
			sort.Slice(keys, func(i, j int) bool {
				return keys[i] < keys[j]
			})
			kvs := make([][2]string, 0, len(keys))
			for _, key := range keys {
				val := os.Getenv(key)
				if strings.HasPrefix(val, "postgres") {
					ep := datatypex.Endpoint{}
					_ = ep.UnmarshalText([]byte(val))
					val = ep.SecurityString()
				}
				if strings.Contains(key, "Private") || strings.Contains(key, "Secret") {
					val = datatypex.MaskedPassword
				}
				kvs = append(kvs, [2]string{key, val})
			}
			c.JSON(http.StatusOK, kvs)
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
	eng.Handle(
		http.MethodGet, "/metrics",
		gin.WrapH(promhttp.Handler()),
	)
	eng.Run(":80")
}
