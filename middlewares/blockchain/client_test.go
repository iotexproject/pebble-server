package blockchain_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/xhd2015/xgo/runtime/mock"

	"github.com/iotexproject/pebble-server/middlewares/blockchain"
)

func TestEthClient_Init(t *testing.T) {
	r := require.New(t)
	c := &blockchain.EthClient{
		Network:  blockchain.NETWORK__IOTX_TESTNET,
		Endpoint: "https://babel-api.testnet.iotex.io",
	}

	t.Run("Success", func(t *testing.T) {
		r.NoError(c.Init())
	})

	t.Run("InvalidNetwork", func(t *testing.T) {
		c.Network = blockchain.NETWORK_UNKNOWN
		r.ErrorContains(c.Init(), "invalid network")
	})

	t.Run("FailedToDailEth", func(t *testing.T) {
		c.Network = blockchain.NETWORK__IOTX_TESTNET
		mock.Patch(ethclient.Dial, func(string) (*ethclient.Client, error) {
			return nil, errors.New(t.Name())
		})
		r.ErrorContains(c.Init(), t.Name())
	})

	t.Run("FailedToGetChainID", func(t *testing.T) {
		c = &blockchain.EthClient{
			Network:  blockchain.NETWORK__IOTX_TESTNET,
			Endpoint: "https://invalid.chain.endpoint.com",
		}
		r.ErrorContains(c.Init(), "failed to get chain id")
	})
}
