package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/xoctopus/confx/confapp"
	"github.com/xoctopus/confx/confmws/confmqtt"
	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/misc/must"

	"github.com/machinefi/sprout-pebble-sequencer/cmd/sequencer/commands"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/blockchain"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/crypto"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/database"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/logger"
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
		MqttBroker     *confmqtt.Broker
		Database       *database.Postgres
		Blockchain     *blockchain.Blockchain
		Logger         *logger.Logger
		PrivateKey     *crypto.EcdsaPrivateKey
		ProjectID      uint64
		ProjectVersion string
		WhiteList      contexts.WhiteList
	}{
		Logger:     &logger.Logger{Level: slog.LevelDebug},
		Blockchain: &blockchain.Blockchain{Contracts: contracts},
		MqttBroker: &confmqtt.Broker{},
		Database:   &database.Postgres{},
		// from sprout default sequencer, to make coordinator validate sequencer signature
		PrivateKey: &crypto.EcdsaPrivateKey{
			Hex: "dbfe03b0406549232b8dccc04be8224fcc0afa300a33d4f335dcfdfead861c85",
		},
		// WhiteList: contexts.WhiteList{"103381234567407"},
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
	)

	app.Conf(config)
	must.BeTrueWrap(
		config.ProjectVersion != "" && config.ProjectID != 0,
		"project id and version is required",
	)

	ctx = contextx.WithContextCompose(
		contexts.WithLoggerContext(config.Logger),
		contexts.WithBlockchainContext(config.Blockchain),
		contexts.WithDatabaseContext(config.Database),
		contexts.WithMqttBrokerContext(config.MqttBroker),
		contexts.WithProjectIDContext(config.ProjectID),
		contexts.WithProjectVersionContext(config.ProjectVersion),
		contexts.WithEcdsaPrivateKeyContext(config.PrivateKey),
		contexts.WithWhiteListKeyContext(config.WhiteList),
	)(context.Background())

	app.AddCommand(commands.Migrate(ctx))
	app.AddCommand(commands.GenerateSproutConfig(ctx))
}

// Main app main entry
func Main() error {
	if err := config.Blockchain.RunMonitors(); err != nil {
		config.Logger.Error(err, "failed to start tx monitor")
	}
	event.InitRunner(ctx)()
	defer config.Blockchain.Close()

	go RunDebugServer(ctx)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	_ = <-sig

	return nil
}

// RunDebugServer enable simple http server for debugging
func RunDebugServer(ctx context.Context) {
	// addr := contexts.ServerAddrFromContext(ctx)
	eng := gin.Default()
	eng.Handle(
		http.MethodGet, "/debug/monitor-info",
		func(c *gin.Context) {
			bc := must.BeTrueV(contexts.BlockchainFromContext(ctx))
			monitors := bc.MonitorsInfo()
			sort.Slice(monitors, func(i, j int) bool {
				return monitors[i].Name < monitors[j].Name
			})

			name := c.Query("name")
			if name == "" {
				c.JSON(http.StatusOK, monitors)
				return
			} else {
				for _, m := range monitors {
					if m.Name == name {
						c.JSON(http.StatusOK, m)
						return
					}
				}
			}
			c.Status(http.StatusNotFound)
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

func main() {
	if err := app.Command.Execute(); err != nil {
		app.PrintErrln(err)
	}
	config.Blockchain.Close()
	os.Exit(-1)
}
