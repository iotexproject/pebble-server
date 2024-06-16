package blockchain_test

import (
	"testing"

	"github.com/cockroachdb/pebble"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xoctopus/x/ptrx"

	. "github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/blockchain"
)

func TestBlockchain_Init(t *testing.T) {
	r := require.New(t)

	contract1 := &Contract{
		ID:      "SproutProjectRegistrar",
		Network: NETWORK__IOTX_TESTNET,
		Address: common.HexToAddress("0x4888bfbf39Dc83C19cbBcb307ccE8F7F93b72E38"),
		Events: []*Event{
			{
				Name: "ProjectRegistered",
				ABI:  `[{"anonymous":false,"inputs":[{"indexed":true,"internalType":"uint256","name":"projectId","type":"uint256"}],"name":"ProjectRegistered","type":"event"}]`,
			},
		},
	}

	t.Run("InitClients", func(t *testing.T) {
		t.Run("FailedToInitClient", func(t *testing.T) {
			bc := &Blockchain{
				Clients: []*EthClient{{Network: NETWORK_UNKNOWN}},
			}
			bc.SetDefault()
			r.ErrorContains(bc.Init(), "invalid network")
		})
		t.Run("ClientDuplicated", func(t *testing.T) {
			bc := &Blockchain{
				Clients: []*EthClient{
					{Network: NETWORK__IOTX_TESTNET, Endpoint: "https://babel-api.testnet.iotex.io"},
					{Network: NETWORK__IOTX_TESTNET, Endpoint: "https://babel-api.testnet.iotex.io"},
				},
			}
			bc.SetDefault()
			r.ErrorContains(bc.Init(), "client duplicated")
		})
	})

	t.Run("InitContracts", func(t *testing.T) {
		t.Run("NetworkNotFound", func(t *testing.T) {
			contract := ptrx.Ptr(*contract1)
			contract.Network = Network(10) // network not configured

			bc := &Blockchain{Contracts: []*Contract{contract}}
			bc.SetDefault()
			r.ErrorContains(bc.Init(), "contract network `10` not found")
		})
		t.Run("FailedToInitContract", func(t *testing.T) {
			contract := ptrx.Ptr(*contract1)
			contract.ID = ""

			bc := &Blockchain{Contracts: []*Contract{contract}}
			bc.SetDefault()
			r.ErrorContains(bc.Init(), "failed to init contract")
		})
		t.Run("ContractIDConflict", func(t *testing.T) {
			bc := &Blockchain{Contracts: []*Contract{ptrx.Ptr(*contract1), ptrx.Ptr(*contract1)}}
			bc.SetDefault()
			r.ErrorContains(bc.Init(), "contract id `SproutProjectRegistrar` duplicated")
		})
		t.Run("ContractAddressOrNetworkConflict", func(t *testing.T) {
			bc := &Blockchain{Contracts: []*Contract{ptrx.Ptr(*contract1), ptrx.Ptr(*contract1)}}
			bc.Contracts[0].ID = "any"
			bc.SetDefault()
			r.ErrorContains(bc.Init(), "contract `SproutProjectRegistrar` duplicated with `any`")
		})
	})

	t.Run("InitPersistence", func(t *testing.T) {
		t.Run("FailedToInitPersistence", func(t *testing.T) {
			mock.Patch(pebble.Open, func(string, *pebble.Options) (*pebble.DB, error) {
				return nil, errors.New(t.Name())
			})
			bc := &Blockchain{
				PersistPath: dir(t),
				Contracts:   []*Contract{ptrx.Ptr(*contract1)},
			}
			bc.SetDefault()
			err := bc.Init()
			r.ErrorContains(err, "failed to init persistence")
			r.ErrorContains(err, t.Name())
		})
	})

	t.Run("InitMonitors", func(t *testing.T) {
		t.Run("FailedToInitMonitor", func(t *testing.T) {
			// mock a invalid meta range to make persist.Init failed
			topic := common.BytesToHash(crypto.Keccak256([]byte("ProjectRegistered(uint256)")))
			p := &Persist{Path: dir(t)}
			r.NoError(p.Init())
			r.NoError(p.Store(MetaRangeEndKey((&Meta{
				Network:  contract1.Network,
				Contract: contract1.Address,
				Topic:    topic,
			}).MetaID()), make([]byte, 10)))
			r.NoError(p.Close())

			bc := &Blockchain{
				PersistPath: p.Path,
				Contracts:   []*Contract{ptrx.Ptr(*contract1)},
			}
			bc.SetDefault()

			err := bc.Init()
			r.ErrorContains(err, "failed to init monitor")
		})
	})

	bc := &Blockchain{
		PersistPath: dir(t),
		Contracts:   []*Contract{ptrx.Ptr(*contract1)},
	}
	bc.SetDefault()

	t.Run("Success", func(t *testing.T) {
		r.NoError(bc.Init())
	})

	t.Run("Ext", func(t *testing.T) {
		r.NotNil(bc.ClientByNetwork(NETWORK__IOTX_MAINNET))
		r.NotNil(bc.ClientByNetwork(NETWORK__IOTX_TESTNET))
		r.Nil(bc.ClientByNetwork(100))

		r.NotNil(bc.ContractByID("SproutProjectRegistrar"))

		r.NotNil(bc.Monitor("SproutProjectRegistrar", "ProjectRegistered"))
		r.Nil(bc.Monitor("not-found-contract", "ProjectRegistered"))
		r.Nil(bc.Monitor("SproutProjectRegistrar", "not-found-event"))
	})
}
