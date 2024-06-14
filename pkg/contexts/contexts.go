package contexts

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-logr/logr"
	"github.com/xoctopus/confx/confmws/confmqtt"
	"github.com/xoctopus/x/contextx"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/blockchain"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/database"
)

type (
	ctxLogger     struct{}
	ctxMqttBroker struct{}
	ctxMqttClient struct{}
	ctxBlockchain struct{}
	ctxDatabase   struct{}
	ctxContracts  struct{}
)

func LoggerFromContext(ctx context.Context) (*logr.Logger, bool) {
	v, ok := ctx.Value(ctxLogger{}).(*logr.Logger)
	return v, ok
}

func WithLoggerContext(v *logr.Logger) contextx.WithContext {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, ctxLogger{}, v)
	}
}

func MqttBrokerFromContext(ctx context.Context) (*confmqtt.Broker, bool) {
	v, ok := ctx.Value(ctxMqttBroker{}).(*confmqtt.Broker)
	return v, ok
}

func WithMqttBrokerContext(v *confmqtt.Broker) contextx.WithContext {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, ctxMqttBroker{}, v)
	}
}

func MqttClientFromContext(ctx context.Context) (*confmqtt.Client, bool) {
	v, ok := ctx.Value(ctxMqttClient{}).(*confmqtt.Client)
	return v, ok
}

func WithMqttClientContext(v *confmqtt.Client) contextx.WithContext {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, ctxMqttClient{}, v)
	}
}

func EthClientFromContextByName(ctx context.Context, name string) (*blockchain.EthClient, bool) {
	v, ok := ctx.Value(ctxBlockchain{}).(blockchain.Blockchain)
	if !ok {
		return nil, false
	}
	c := v.ClientByName(name)
	return c, c != nil
}

func EthClientFromContextByEndpoint(ctx context.Context, endpoint string) (*blockchain.EthClient, bool) {
	v, ok := ctx.Value(ctxBlockchain{}).(blockchain.Blockchain)
	if !ok {
		return nil, false
	}
	c := v.ClientByEndpoint(endpoint)
	return c, c != nil
}

func BlockchainFromContext(ctx context.Context) (*blockchain.Blockchain, bool) {
	v, ok := ctx.Value(ctxBlockchain{}).(*blockchain.Blockchain)
	return v, ok
}

func WithBlockchainContext(v *blockchain.Blockchain) contextx.WithContext {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, ctxBlockchain{}, v)
	}
}

func DatabaseFromContext(ctx context.Context) (*database.Postgres, bool) {
	v, ok := ctx.Value(ctxDatabase{}).(*database.Postgres)
	return v, ok
}

func WithDatabaseContext(v *database.Postgres) contextx.WithContext {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, ctxDatabase{}, v)
	}
}

func ContractAddrFromContextByName(ctx context.Context, name string) (common.Address, bool) {
	v, ok := ctx.Value(ctxContracts{}).(map[string]common.Address)
	if ok {
		if addr, ok := v[name]; ok {
			return addr, true
		}
	}
	return common.Address{}, false
}

func WithContractsContext(v map[string]common.Address) contextx.WithContext {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, ctxContracts{}, v)
	}
}
