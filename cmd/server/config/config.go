package config

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/viper"
)

type Config struct {
	LogLevel                 slog.Level `env:"LOG_LEVEL,optional"`
	ServiceEndpoint          string     `env:"HTTP_SERVICE_ENDPOINT"`
	PrvKey                   string     `env:"PRIVATE_KEY,optional"`
	DatabaseDSN              string     `env:"DATABASE_DSN"`
	OldDatabaseDSN           string     `env:"OLD_DATABASE_DSN"`
	ChainEndpoint            string     `env:"CHAIN_ENDPOINT,optional"`
	BeginningBlockNumber     uint64     `env:"BEGINNING_BLOCK_NUMBER,optional"`
	IoIDProjectID            uint64     `env:"IOID_PROJECT_ID,optional"`
	IoIDRegistryContractAddr string     `env:"IOID_REGISTRY_CONTRACT_ADDRESS,optional"`
	IoIDContractAddr         string     `env:"IOID_CONTRACT_ADDRESS,optional"`
	ProjectContractAddr      string     `env:"PROJECT_CONTRACT_ADDRESS,optional"`
	W3bstreamProjectID       string     `env:"W3BSTREAM_PROJECT_ID,optional"`
	W3bstreamServiceEndpoint string     `env:"W3BSTREAM_SERVICE_ENDPOINT,optional"`
	env                      string     `env:"-"`
}

var (
	defaultTestnetConfig = &Config{
		LogLevel:                 slog.LevelInfo,
		ServiceEndpoint:          ":9000",
		DatabaseDSN:              "postgres://postgres:mysecretpassword@postgres:5432/w3bstream?sslmode=disable",
		ChainEndpoint:            "https://babel-api.testnet.iotex.io",
		BeginningBlockNumber:     28685000,
		IoIDProjectID:            915,
		IoIDRegistryContractAddr: "0x0A7e595C7889dF3652A19aF52C18377bF17e027D",
		IoIDContractAddr:         "0x45Ce3E6f526e597628c73B731a3e9Af7Fc32f5b7",
		ProjectContractAddr:      "0xf07336E1c77319B4e740b666eb0C2B19D11fc14F",
		env:                      "TESTNET",
	}
	defaultMainnetConfig = &Config{
		LogLevel:                 slog.LevelInfo,
		ServiceEndpoint:          ":9000",
		DatabaseDSN:              "postgres://postgres:mysecretpassword@postgres:5432/w3bstream?sslmode=disable",
		ChainEndpoint:            "https://babel-api.mainnet.iotex.io",
		BeginningBlockNumber:     28685000,
		IoIDProjectID:            6,
		IoIDRegistryContractAddr: "0x04e4655Cf258EC802D17c23ec6112Ef7d97Fa2aF",
		IoIDContractAddr:         "0x1FCB980eD0287777ab05ADc93012332e11300e54",
		ProjectContractAddr:      "0xA596800891e6a95Bf737404411ef529c1F377b4e",
		env:                      "MAINNET",
	}
)

func (c *Config) init() error {
	if err := parseEnv(c); err != nil {
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
	case "MAINNET":
		conf = defaultMainnetConfig
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
	print(c)
}

func parseEnvTag(tag string) (key string, require bool) {
	if tag == "" || tag == "-" {
		return "", false
	}
	tagKeys := strings.Split(tag, ",")
	key = tagKeys[0]
	if len(tagKeys) > 1 && tagKeys[1] == "optional" {
		return key, false
	}
	return key, true
}

func parseEnv(c any) error {
	rv := reflect.ValueOf(c).Elem()
	rt := reflect.TypeOf(c).Elem()

	for i := 0; i < rt.NumField(); i++ {
		fi := rt.Field(i)
		fv := rv.Field(i)
		key, require := parseEnvTag(fi.Tag.Get("env"))
		if key == "" {
			continue
		}
		viper.MustBindEnv(key)

		v := viper.Get(key)
		if require && v == nil && fv.IsZero() {
			panic(fmt.Sprintf("env `%s` is require but got empty", key))
		}
		if v == nil {
			continue
		}

		switch fv.Kind() {
		case reflect.String:
			fv.Set(reflect.ValueOf(viper.GetString(key)))
		case reflect.Int:
			if fi.Type == reflect.TypeOf(slog.Level(0)) {
				level := slog.Level(viper.GetInt(key))
				fv.Set(reflect.ValueOf(level))
			} else {
				fv.Set(reflect.ValueOf(viper.GetInt(key)))
			}
		case reflect.Uint64:
			fv.Set(reflect.ValueOf(viper.GetUint64(key)))
		}
	}
	return nil
}

func print(c any) {
	rt := reflect.TypeOf(c).Elem()
	rv := reflect.ValueOf(c).Elem()

	if env, ok := c.(interface{ Env() string }); ok {
		fmt.Println(color.CyanString("ENV: %s", env.Env()))
	}

	for i := 0; i < rt.NumField(); i++ {
		fi := rt.Field(i)
		fv := rv.Field(i)
		key, _ := parseEnvTag(fi.Tag.Get("env"))
		if key == "" {
			continue
		}
		fmt.Printf("%s: %v\n", color.GreenString(key), fv.Interface())
	}
}
