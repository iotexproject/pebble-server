package blockchain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
	"github.com/xoctopus/x/misc/must"
)

type EthClient struct {
	Endpoint string
	Network  Network
	chainID  *big.Int

	*ethclient.Client `evn:"-"`
}

func (c *EthClient) Init() error {
	if c.Network == NETWORK_UNKNOWN || c.Endpoint == "" {
		return errors.Errorf(
			"invalid network or endpoint: [%d] [%s]",
			c.Network, c.Endpoint,
		)
	}

	client, err := ethclient.Dial(c.Endpoint)
	if err != nil {
		return err
	}
	c.Client = client

	c.chainID, err = c.ChainID(context.Background())
	if err != nil {
		return err
	}
	must.BeTrueWrap(
		c.chainID.Int64() == int64(c.Network),
		"unmatched network id: [client:%d] [network:%d]",
		c.chainID, c.Network.Int(),
	)
	return nil
}

func (c *EthClient) ChainEndpoint() string {
	return c.Endpoint
}

func (c *EthClient) ChainID(ctx context.Context) (chainID *big.Int, err error) {
	if c.chainID == nil {
		chainID, err = c.Client.ChainID(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get chain id")
		}
		c.chainID = chainID
	}
	return c.chainID, nil
}
