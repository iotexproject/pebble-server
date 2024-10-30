package blockchain_test

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/iotexproject/pebble-server/middlewares/blockchain"
)

func TestMonitorMeta(t *testing.T) {
	var (
		network   = blockchain.NETWORK__IOTX_TESTNET
		contracts = []*blockchain.Contract{
			{
				ID:      "any",
				Network: blockchain.NETWORK__IOTX_TESTNET,
				Address: common.HexToAddress("0x6AfCB0EB71B7246A68Bb9c0bFbe5cD7c11c4839f"),
				Events:  []*blockchain.Event{{Name: "ProjectConfigUpdated", ABI: ProjectConfigUpdatedABI}},
			},
			{
				ID:      "any2",
				Network: blockchain.NETWORK__IOTX_MAINNET,
			},
		}
	)
	r := require.New(t)

	r.NoError(contracts[0].Init())

	meta := blockchain.NewMeta(network, contracts...)

	r.Len(meta.Contracts, 1)

	s := meta.String()
	lk := string(meta.MonitorRangeFromKey())
	r.Equal(lk, "RNL_"+s)
	hk := string(meta.MonitorRangeEndKey())
	r.Equal(hk, "RNH_"+s)
	t.Log(string(meta.WatcherRangeFromKey("abc")))
	t.Log(string(meta.WatcherRangeEndKey("abc")))
	t.Log(string(meta.BlockKeyPrefixLowerBound(100)))
	t.Log(string(meta.BlockKeyPrefixUpperBound(100)))
	t.Log(string(meta.LogKey(&types.Log{
		Address:     common.HexToAddress("0x0000000000000000000000000000000000000001"),
		Topics:      []common.Hash{common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000002")},
		BlockNumber: 100,
		TxHash:      common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000003"),
	})))

	parts := [7]string{
		"ANY",  // 0 prefix
		"4690", // 1 network
		"",     // 2 tx index
		"",     // 3 block number
		"",     // 4 contract address
		"",     // 5 topic hash
		"",     // 6 tx hash
	}
	_, _, err := meta.ParseLogKey(strings.Join(parts[:5], "_"))
	r.ErrorContains(err, "invalid log key parts")

	_, _, err = meta.ParseLogKey(strings.Join(parts[:], "_"))
	r.ErrorContains(err, "invalid log key prefix")

	parts[0] = "LOG"
	parts[1] = "100"
	_, _, err = meta.ParseLogKey(strings.Join(parts[:], "_"))
	r.ErrorContains(err, "invalid log key network")

	parts[1] = "4690"
	parts[2] = "any"
	_, _, err = meta.ParseLogKey(strings.Join(parts[:], "_"))
	r.ErrorContains(err, "invalid log key block number")

	parts[2] = "100"
	parts[3] = "abc"
	_, _, err = meta.ParseLogKey(strings.Join(parts[:], "_"))
	r.ErrorContains(err, "invalid log key tx index")

	parts[3] = "1"
	_, _, err = meta.ParseLogKey(strings.Join(parts[:], "_"))
	r.ErrorContains(err, "contract not found")

	parts[4] = "0x6AfCB0EB71B7246A68Bb9c0bFbe5cD7c11c4839f"
	_, _, err = meta.ParseLogKey(strings.Join(parts[:], "_"))
	r.ErrorContains(err, "event not found")

	name := "ProjectConfigUpdated"
	topic, _ := contracts[0].Topic(name)
	parts[5] = topic.String()
	c, event, err := meta.ParseLogKey(strings.Join(parts[:], "_"))
	r.NoError(err)
	r.NotNil(c)
	r.Equal(event, name)
}
