package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/xoctopus/confx/confapp"
	"github.com/xoctopus/confx/confmws/confmqtt"
	"github.com/xoctopus/x/misc/must"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/logger"
)

var (
	Name     = "simulator"
	Feature  string
	Version  string
	CommitID string
	Date     string

	app    *confapp.AppCtx
	config = &struct {
		MqttBroker *confmqtt.Broker
		Logger     *logger.Logger
		Devices    []string
	}{
		MqttBroker: &confmqtt.Broker{},
		Logger:     &logger.Logger{Level: slog.LevelDebug},
	}
)

func init() {
	meta := confapp.Meta{
		Name:     Name,
		Feature:  Feature,
		Version:  Version,
		CommitID: CommitID,
		Date:     Date,
	}
	app = confapp.NewAppContext(
		confapp.WithBuildMeta(meta),
		confapp.WithMainRoot("."),
		confapp.WithMainExecutor(Main),
	)

	app.Conf(config)
}

func Main() error {
	if len(config.Devices) == 0 {
		return nil
	}

	clients := make([]*confmqtt.Client, 0, len(config.Devices)*2)

	for _, imei := range config.Devices {
		clients = append(clients, PubSubQuery(imei)...)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	_ = <-sig

	for _, c := range clients {
		config.Logger.Info("client closing", "client", c.ID())
		config.MqttBroker.Close(c)
	}

	return nil
}

func main() {
	if err := app.Command.Execute(); err != nil {
		app.PrintErrln(err)
	}
	os.Exit(-1)
}

func PubSubQuery(imei string) []*confmqtt.Client {
	broker := config.MqttBroker
	logger := config.Logger
	clients := make([]*confmqtt.Client, 2)

	{
		topic := "backend/+/status"
		client, err := broker.NewClient("sub_backend_status_simulator", topic)
		must.NoErrorWrap(err, "failed to new sub mqtt client: [topic %s]", topic)
		clients[1] = client
		logger.Info("subscribing started", "client", client.ID(), "topic", topic)
		sequence := 0
		err = client.Subscribe(func(_ mqtt.Client, message mqtt.Message) {
			if parts := strings.Split(message.Topic(), "/"); len(parts) != 3 || parts[1] != imei {
				return
			}

			rsp := &struct {
				Status     int32  `json:"status"`
				Proposer   string `json:"proposer,omitempty"`
				Firmware   string `json:"firmware,omitempty"`
				URI        string `json:"uri,omitempty"`
				Version    string `json:"version,omitempty"`
				ServerMeta string `json:"server_meta"`
			}{}
			pl := message.Payload()
			if err = json.Unmarshal(pl, rsp); err != nil {
				logger.Error(err, "failed to unmarshal response", "seq", sequence, "topic", topic, "response", string(pl))
			} else {
				logger.Info("sub", "seq", sequence, "data", rsp, "topic", topic)
			}
			sequence++
		})
		if err != nil {
			logger.Error(err, "failed to subscribing", "topic", topic)
			panic(err)
		}
	}

	{
		topic := "device/" + imei + "/query"
		client, err := broker.NewClient(imei+"_"+uuid.NewString(), topic)
		must.NoErrorWrap(err, "failed to new pub mqtt client: [topic %s]", topic)
		clients[0] = client
		logger.Info("publishing started", "client", client.ID(), "topic", topic)
		go func() {
			sequence := 0
			for {
				err := client.Publish([]byte{})
				if err != nil {
					logger.Error(err, "failed to publish", "seq", sequence, "topic", topic)
				} else {
					logger.Info("pub", "seq", sequence, "topic", topic)
					sequence++
				}
				time.Sleep(time.Second * 5)
			}
		}()
	}

	return clients
}
