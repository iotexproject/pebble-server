package blockchain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/iotexproject/pebble-server/middlewares/blockchain"
)

func TestNetwork(t *testing.T) {
	r := require.New(t)

	for _, s := range []string{"", "IOTX_MAINNET", "IOTX_TESTNET", "Invalid"} {
		network := Network(0)
		if err := network.UnmarshalText([]byte(s)); err != nil {
			r.Equal(err, InvalidNetwork)
		} else {
			r.Equal(network.String(), s)
		}
	}

	for _, network := range []Network{
		NETWORK_UNKNOWN,
		NETWORK__IOTX_MAINNET,
		NETWORK__IOTX_TESTNET,
		Network(-1), //invalid
	} {
		s, err := network.MarshalText()
		if err != nil {
			r.Equal(err, InvalidNetwork)
			r.Nil(s)
		} else {
			r.Equal(network.String(), string(s))
			r.EqualValues(network, network.Int())
		}
	}
}
