package blockchain

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

type EthClient struct {
	Endpoint string

	*ethclient.Client `evn:"-"`
}

func (c *EthClient) Init() error {
	client, err := ethclient.Dial(c.Endpoint)
	if err != nil {
		return err
	}
	c.Client = client
	return nil
}

type EthClients map[string]*EthClient

func (clients EthClients) Init() error {
	for name, client := range clients {
		if err := client.Init(); err != nil {
			return errors.Wrapf(err, "failed to dail eth client: %s %s", name, client.Endpoint)
		}
	}
	return nil
}
