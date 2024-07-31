package contexts

import (
	"github.com/xoctopus/confx/confapp"
	"github.com/xoctopus/confx/confmws/confmqtt"
	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/ptrx"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/alert"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/blockchain"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/crypto"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/database"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/logger"
)

var (
	_dryRun         = contextx.NewValue(false)
	_appMeta        = contextx.NewValue(ptrx.Ptr(confapp.DefaultMeta))
	_larkAlert      = contextx.New[*alert.LarkAlert]()
	_whiteList      = contextx.New[WhiteList]()
	_privateKey     = contextx.New[*crypto.EcdsaPrivateKey]()
	_projectID      = contextx.New[uint64]()
	_projectVersion = contextx.New[string]()
	_database       = contextx.New[*database.Postgres]()
	_blockchain     = contextx.New[*blockchain.Blockchain]()
	_mqttBroker     = contextx.New[*confmqtt.Broker]()
	_logger         = contextx.New[*logger.Logger]()
	_mqttClientID   = contextx.New[string]()
)

func DryRun() contextx.Context[bool]                        { return _dryRun }
func AppMeta() contextx.Context[*confapp.Meta]              { return _appMeta }
func LarkAlert() contextx.Context[*alert.LarkAlert]         { return _larkAlert }
func IMEIFilter() contextx.Context[WhiteList]               { return _whiteList }
func PrivateKey() contextx.Context[*crypto.EcdsaPrivateKey] { return _privateKey }
func ProjectID() contextx.Context[uint64]                   { return _projectID }
func ProjectVersion() contextx.Context[string]              { return _projectVersion }
func Database() contextx.Context[*database.Postgres]        { return _database }
func Blockchain() contextx.Context[*blockchain.Blockchain]  { return _blockchain }
func MqttBroker() contextx.Context[*confmqtt.Broker]        { return _mqttBroker }
func Logger() contextx.Context[*logger.Logger]              { return _logger }
func MqttClientID() contextx.Context[string]                { return _mqttClientID }
