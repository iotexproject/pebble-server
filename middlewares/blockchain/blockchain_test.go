package blockchain_test

import (
	"testing"

	"github.com/cockroachdb/pebble"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/xhd2015/xgo/runtime/mock"
	"github.com/xoctopus/x/ptrx"

	. "github.com/iotexproject/pebble-server/middlewares/blockchain"
)

func TestBlockchain_Init(t *testing.T) {
	r := require.New(t)

	contract := &Contract{
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
				Clients: []*EthClient{{Network: NETWORK__IOTX_MAINNET}},
			}
			bc.SetDefault()
			r.ErrorContains(bc.Init(), "invalid network")
		})
		t.Run("ClientDuplicated", func(t *testing.T) {
			bc := &Blockchain{
				Clients: []*EthClient{
					{Network: NETWORK__IOTX_TESTNET, Endpoint: "https://babel-api.testnet.iotex.io"},
					{Network: NETWORK__IOTX_MAINNET, Endpoint: "https://babel-api.mainnet.iotex.io"},
					{Network: NETWORK__IOTX_MAINNET, Endpoint: "https://babel-api.mainnet.iotex.io"},
				},
			}
			bc.SetDefault()
			r.ErrorContains(bc.Init(), "client duplicated")
		})
	})

	t.Run("InitContracts", func(t *testing.T) {
		t.Run("NetworkNotFound", func(t *testing.T) {
			contract := ptrx.Ptr(*contract)
			contract.Network = NETWORK__IOTX_MAINNET
			contract.Address = common.Address{}

			bc := &Blockchain{Contracts: []*Contract{contract}}
			bc.SetDefault()
			r.ErrorContains(bc.Init(), "invalid contract")
		})
		t.Run("FailedToInitContract", func(t *testing.T) {
			contract := ptrx.Ptr(*contract)
			contract.ID = ""
			contract.Network = NETWORK__IOTX_MAINNET

			bc := &Blockchain{Contracts: []*Contract{contract}}
			bc.SetDefault()
			r.ErrorContains(bc.Init(), "failed to init contract")
		})
		t.Run("ContractIDConflict", func(t *testing.T) {
			bc := &Blockchain{
				Network:   NETWORK__IOTX_TESTNET,
				Contracts: []*Contract{ptrx.Ptr(*contract), ptrx.Ptr(*contract)},
			}
			bc.SetDefault()

			r.ErrorContains(bc.Init(), "contract id `SproutProjectRegistrar` duplicated")
		})
		t.Run("ContractAddressOrNetworkConflict", func(t *testing.T) {
			bc := &Blockchain{
				Network:   NETWORK__IOTX_TESTNET,
				Contracts: []*Contract{ptrx.Ptr(*contract), ptrx.Ptr(*contract)},
			}
			bc.Contracts[0].ID = "any"
			bc.SetDefault()
			r.ErrorContains(bc.Init(), "contract `any` duplicated with `SproutProjectRegistrar`")
		})
	})

	t.Run("InitPersistence", func(t *testing.T) {
		t.Run("FailedToInitPersistence", func(t *testing.T) {
			mock.Patch(pebble.Open, func(string, *pebble.Options) (*pebble.DB, error) {
				return nil, errors.New(t.Name())
			})
			bc := &Blockchain{
				PersistPath: dir(t),
				Contracts:   []*Contract{ptrx.Ptr(*contract)},
			}
			bc.SetDefault()
			err := bc.Init()
			r.ErrorContains(err, "failed to init bc persistence")
			r.ErrorContains(err, t.Name())
		})
	})

	t.Run("InitMonitors", func(t *testing.T) {
		t.Run("FailedToInitMonitor", func(t *testing.T) {
			// mock a invalid meta range to make persist.Init failed
			c2 := &Contract{
				ID:      contract.ID,
				Network: contract.Network,
				Address: contract.Address,
				Events:  []*Event{ptrx.Ptr(*contract.Events[0])},
			}
			r.NoError(c2.Init())
			meta := NewMeta(contract.Network, c2)
			p := &Persist{Path: dir(t)}
			r.NoError(p.Init())
			r.NoError(p.Store(meta.MonitorRangeEndKey(), make([]byte, 10)))
			r.NoError(p.Close())

			bc := &Blockchain{
				Network:     contract.Network,
				PersistPath: p.Path,
				Contracts:   []*Contract{ptrx.Ptr(*contract)},
			}
			bc.SetDefault()
			err := bc.Init()
			r.NoError(err)
			defer bc.Close()

			_, err = bc.Watch(nil, nil)
			r.ErrorContains(err, "failed to load monitor range")
		})
	})

	bc := &Blockchain{
		Network:     contract.Network,
		PersistPath: dir(t),
		Contracts:   []*Contract{ptrx.Ptr(*contract)},
	}
	bc.SetDefault()

	t.Run("Success", func(t *testing.T) {
		r.NoError(bc.Init())
		defer bc.Close()
		r.NoError(bc.RunMonitor())
		_, err := bc.Watch(nil, nil)
		r.NoError(err)

		r.NotNil(bc.ClientByNetwork())
		r.NotNil(bc.ContractByID("SproutProjectRegistrar"))
		monitors := bc.MonitorMeta()
		t.Log(monitors)
	})
}
