package config

import (
	"log/slog"
	"os"

	"github.com/iotexproject/pebble-server/util/env"
)

type Config struct {
	LogLevel                     slog.Level `env:"LOG_LEVEL,optional"`
	DatabaseDSN                  string     `env:"DATABASE_DSN"`
	ChainEndpoint                string     `env:"CHAIN_ENDPOINT,optional"`
	BeginningBlockNumber         uint64     `env:"BEGINNING_BLOCK_NUMBER,optional"`
	OperatorPrvKey               string     `env:"OPERATOR_PRIVATE_KEY,optional"`
	LocalDBPath                  string     `env:"LOCAL_DB_PATH,optional"`
	MqttBrokerEndpoint           string     `env:"MQTT_BROKER_ENDPOINT,optional"`
	MqttBrokerQoS                string     `env:"MQTT_BROKER_QOS,optional"`
	MqttBrokerCertCAPath         string     `env:"MQTT_BROKER_CERT_CA_PATH,optional"`
	MqttBrokerCertCrtPath        string     `env:"MQTT_BROKER_CERT_CRT_PATH,optional"`
	MqttBrokerCertKeyPath        string     `env:"MQTT_BROKER_CERT_KEY_PATH,optional"`
	IoIDProjectID                uint64     `env:"IOID_PROJECT_ID,optional"`
	IoIDRegistryEndpoint         string     `env:"IOID_REGISTRY_ENDPOINT,optional"`
	IoIDRegistryContractAddr     string     `env:"IOID_REGISTRY_CONTRACT_ADDRESS,optional"`
	IoIDContractAddr             string     `env:"IOID_CONTRACT_ADDRESS,optional"`
	ProjectDeviceContractAddr    string     `env:"PROJECT_DEVICE_CONTRACT_ADDRESS,optional"`
	W3bstreamProjectContractAddr string     `env:"W3BSTREAM_PROJECT_CONTRACT_ADDRESS,optional"`
	env                          string     `env:"-"`
}

var (
	defaultTestnetConfig = &Config{
		LogLevel:                     slog.LevelInfo,
		DatabaseDSN:                  "postgres://postgres:mysecretpassword@postgres:5432/w3bstream?sslmode=disable",
		ChainEndpoint:                "https://babel-api.testnet.iotex.io",
		BeginningBlockNumber:         28685000,
		LocalDBPath:                  "./local_db",
		MqttBrokerQoS:                "ONCE",
		MqttBrokerCertCAPath:         "/etc/pebble/root.pem",
		MqttBrokerCertCrtPath:        "/etc/pebble/tls-cert.pem",
		MqttBrokerCertKeyPath:        "/etc/pebble/tls-key.pem",
		IoIDProjectID:                915,
		IoIDRegistryEndpoint:         "did.iotex.me",
		IoIDRegistryContractAddr:     "0x0A7e595C7889dF3652A19aF52C18377bF17e027D",
		IoIDContractAddr:             "0x45Ce3E6f526e597628c73B731a3e9Af7Fc32f5b7",
		ProjectDeviceContractAddr:    "0xF4d6282C5dDD474663eF9e70c927c0d4926d1CEb",
		W3bstreamProjectContractAddr: "0x6AfCB0EB71B7246A68Bb9c0bFbe5cD7c11c4839f",
		env:                          "TESTNET",
	}
)

func (c *Config) init() error {
	if err := env.ParseEnv(c); err != nil {
		return err
	}
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.Level(c.LogLevel)})
	slog.SetDefault(slog.New(h))
	return nil
}

func Get() (*Config, error) {
	var conf *Config
	env := os.Getenv("ENV")
	switch env {
	case "TESTNET":
		conf = defaultTestnetConfig
	default:
		env = "TESTNET"
		conf = defaultTestnetConfig
	}
	conf.env = env
	if err := conf.init(); err != nil {
		return nil, err
	}
	return conf, nil
}

func (c *Config) Print() {
	env.Print(c)
}
