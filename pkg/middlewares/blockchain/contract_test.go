package blockchain_test

import (
	_ "embed"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/blockchain"
)

//go:embed testdata/abi.json
var ProjectConfigUpdatedABI string

func TestContract(t *testing.T) {
	r := require.New(t)
	abi := `[{"anonymous":false,"inputs":[{"indexed":false,"internalType":"address","name":"owner","type":"address"},{"indexed":false,"internalType":"string","name":"name","type":"string"},{"indexed":false,"internalType":"string","name":"avatar","type":"string"}],"name":"Updated","type":"event"}]`

	t.Run("InvalidIDOrNetwork", func(t *testing.T) {
		c := &blockchain.Contract{}
		r.ErrorContains(c.Init(), "invalid contract id or network")

		c.ID = "any"
		r.ErrorContains(c.Init(), "invalid contract id or network")
	})

	t.Run("InitEvent", func(t *testing.T) {
		c := &blockchain.Contract{
			ID:      "any",
			Network: blockchain.NETWORK__IOTX_MAINNET,
			Address: common.Address{},
			Events: []*blockchain.Event{
				{Name: "", ABI: ""},
			},
		}
		t.Run("InvalidABIContentOrEventName", func(t *testing.T) {
			r.ErrorContains(c.Init(), "invalid abi content or event name")
		})
		t.Run("FailedToParseABI", func(t *testing.T) {
			c.Events[0].Name = "any"
			c.Events[0].ABI = "invalid abi content"
			r.ErrorContains(c.Init(), "failed to parse abi content to target")
		})
		t.Run("EventNotFoundInABI", func(t *testing.T) {
			c.Events[0].ABI = abi
			r.ErrorContains(c.Init(), "event `any` not found in target abi")
		})

		t.Run("Success", func(t *testing.T) {
			c.Events[0].Name = "Updated"
			r.NoError(c.Init())

			_, exists := c.Topic("Updated")
			r.True(exists)
			_, exists = c.Topic("not-found")
			r.False(exists)
		})
	})

	t.Run("EventDuplicated", func(t *testing.T) {
		c := &blockchain.Contract{
			ID:      "any",
			Network: blockchain.NETWORK__IOTX_MAINNET,
			Address: common.Address{},
			Events: []*blockchain.Event{
				{Name: "Updated", ABI: abi},
				{Name: "Updated", ABI: abi},
			},
		}
		r.ErrorContains(c.Init(), "event `Updated` duplicated")
	})

	t.Run("ParseLog", func(t *testing.T) {
		c := &blockchain.Contract{
			ID:      "any",
			Network: blockchain.NETWORK__IOTX_TESTNET,
			Address: common.HexToAddress("0x6AfCB0EB71B7246A68Bb9c0bFbe5cD7c11c4839f"),
			Events:  []*blockchain.Event{{Name: "ProjectConfigUpdated", ABI: ProjectConfigUpdatedABI}},
		}
		r.NoError(c.Init())

		// this tx was build by https://testnet.iotexscan.io/tx/84d60de8a2953aa46b0039e296714c941be46b52e65f2d903a374edf4c31d212
		var (
			address = common.HexToAddress("0x6AfCB0EB71B7246A68Bb9c0bFbe5cD7c11c4839f")
			topics  = []common.Hash{
				common.HexToHash("0xa9ee0c223bc138bec6ebb21e09d00d5423fc3bbc210bdb6aef9d190b0641aecb"),
				common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000017"),
			}
			hash    = common.HexToHash("0x84d60de8a2953aa46b0039e296714c941be46b52e65f2d903a374edf4c31d212")
			data, _ = hex.DecodeString("00000000000000000000000000000000000000000000000000000000000000401975829605f6c10cab41ffeece8b632090f1a6d47f7cb5d03336ae39635016b9000000000000000000000000000000000000000000000000000000000000004b697066733a2f2f697066732e6d61696e6e65742e696f7465782e696f2f516d6431354e46374453754452345635346833467156457a6279783236677934357438376d79484b63336f7a3145000000000000000000000000000000000000000000")
		)

		type result struct {
			ProjectId *big.Int
			Uri       string
			Hash      common.Hash
		}

		log := &types.Log{
			Address:     address,
			Topics:      topics,
			Data:        data,
			BlockNumber: 259540330,
			TxHash:      hash,
		}

		t.Run("InvalidAddress", func(t *testing.T) {
			log.Address = common.HexToAddress("mismatched")
			r.ErrorContains(c.ParseTxLog("any", log, &result{}), "log address mismatched")
		})

		log.Address = address
		t.Run("EventABINotFound", func(t *testing.T) {
			r.ErrorContains(c.ParseTxLog("mismatch", log, &result{}), "event abi not found")
		})

		t.Run("EventSigNotFound", func(t *testing.T) {
			log.Topics = nil
			r.ErrorContains(c.ParseTxLog("ProjectConfigUpdated", log, &result{}), "event signature not found")
		})

		t.Run("EventSigMismatched", func(t *testing.T) {
			log.Topics = topics[1:]
			r.ErrorContains(c.ParseTxLog("ProjectConfigUpdated", log, &result{}), "event signature mismatched")
		})

		log.Topics = topics
		t.Run("FailedToUnpackTarget", func(t *testing.T) {
			log.Data = []byte("invalid")
			r.ErrorContains(c.ParseTxLog("ProjectConfigUpdated", log, &result{}), "failed to unpack")
		})

		log.Data = data
		t.Run("PanicWhenUnpack", func(t *testing.T) {
			r.Error(c.ParseTxLog("ProjectConfigUpdated", log, (*result)(nil)))
		})
		t.Run("Success", func(t *testing.T) {
			res := &result{}
			r.NoError(c.ParseTxLog("ProjectConfigUpdated", log, res))
			r.Equal(res.ProjectId.Int64(), int64(23))
			r.Equal(res.Uri, "ipfs://ipfs.mainnet.iotex.io/Qmd15NF7DSuDR4V54h3FqVEzbyx26gy45t87myHKc3oz1E")
			r.Equal(res.Hash.String(), "0x1975829605f6c10cab41ffeece8b632090f1a6d47f7cb5d03336ae39635016b9")
		})

		t.Run("HasRedundantFailed", func(t *testing.T) {
			v1 := &struct {
				ProjectId  *big.Int
				Uri        string
				Hash       common.Hash
				unexported int
			}{}
			r.NoError(c.ParseTxLog("ProjectConfigUpdated", log, v1))
			r.Equal(v1.ProjectId.Int64(), int64(23))
			r.Equal(v1.Uri, "ipfs://ipfs.mainnet.iotex.io/Qmd15NF7DSuDR4V54h3FqVEzbyx26gy45t87myHKc3oz1E")
			r.Equal(v1.Hash.String(), "0x1975829605f6c10cab41ffeece8b632090f1a6d47f7cb5d03336ae39635016b9")

			v2 := &struct {
				ProjectId *big.Int
				Uri       string
				Hash      common.Hash
				Exported  int
			}{}
			r.NoError(c.ParseTxLog("ProjectConfigUpdated", log, v2))
			r.Equal(v2.ProjectId.Int64(), int64(23))
			r.Equal(v2.Uri, "ipfs://ipfs.mainnet.iotex.io/Qmd15NF7DSuDR4V54h3FqVEzbyx26gy45t87myHKc3oz1E")
			r.Equal(v2.Hash.String(), "0x1975829605f6c10cab41ffeece8b632090f1a6d47f7cb5d03336ae39635016b9")

			v3 := &struct {
				ProjectId *big.Int
				Uri       string
				Disturbed int
				Hash      common.Hash
			}{}
			r.NoError(c.ParseTxLog("ProjectConfigUpdated", log, v3))
			r.Equal(v3.ProjectId.Int64(), int64(23))
			r.Equal(v3.Uri, "ipfs://ipfs.mainnet.iotex.io/Qmd15NF7DSuDR4V54h3FqVEzbyx26gy45t87myHKc3oz1E")
			r.Equal(v3.Hash.String(), "0x1975829605f6c10cab41ffeece8b632090f1a6d47f7cb5d03336ae39635016b9")

			v4 := &struct {
				Hash      common.Hash
				Disturbed int
				Uri       string
				ProjectId *big.Int
			}{}
			r.NoError(c.ParseTxLog("ProjectConfigUpdated", log, v4))
			r.Equal(v4.ProjectId.Int64(), int64(23))
			r.Equal(v4.Uri, "ipfs://ipfs.mainnet.iotex.io/Qmd15NF7DSuDR4V54h3FqVEzbyx26gy45t87myHKc3oz1E")
			r.Equal(v4.Hash.String(), "0x1975829605f6c10cab41ffeece8b632090f1a6d47f7cb5d03336ae39635016b9")

			v5 := &struct {
				HASH      common.Hash
				Disturbed int
				URI       string
				ProjectID *big.Int
			}{}
			r.ErrorContains(c.ParseTxLog("ProjectConfigUpdated", log, v5), "can't be found in the given value")
		})
	})
}
