package blockchain_test

import (
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/blockchain"
)

func TestMonitorMeta(t *testing.T) {
	var (
		network  = [4]byte{1, 2, 3, 4}
		contract = common.MaxAddress
		topic    = common.Hash{}
		ser      = append(network[:], append(contract[:], topic[:]...)...)
	)

	r := require.New(t)

	t.Run("AssertMonitorMetaIDLength", func(t *testing.T) {
		r.Equal(len(network)+len(contract)+len(topic), blockchain.MetaIDLength)
		r.Equal(len(ser), blockchain.MetaIDLength)
	})

	t.Run("ParseMonitorMeta", func(t *testing.T) {
		mm, err := blockchain.ParseMonitorMeta(ser)
		r.NoError(err)
		r.Equal(mm.Contract, contract)
		r.Equal(mm.Topic, topic)
		r.Equal(mm.MetaID().Bytes(), ser)
		r.Equal(mm.Bytes(), ser)
		r.Equal(mm.MetaID().String(), hex.EncodeToString(ser))

		t.Run("ComparingMarshalResult", func(t *testing.T) {
			ser2, err := mm.MarshalText()
			r.NoError(err)
			r.Equal(ser, ser2)
		})

		t.Run("InvalidParseDataLength", func(t *testing.T) {
			mm, err = blockchain.ParseMonitorMeta(ser[:blockchain.MetaIDLength-1])
			r.Nil(mm)
			r.Error(err)
		})
	})
}
