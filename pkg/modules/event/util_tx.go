package event

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/blockchain"
)

type CanSetTxHash interface {
	SetTxHash(h common.Hash)
}

type TxHash struct {
	hash common.Hash
}

func (h *TxHash) SetTxHash(hash common.Hash) {
	h.hash = hash
}

func (h TxHash) Hash() common.Hash {
	return h.hash
}

func (h TxHash) String() string {
	return h.hash.String()
}

type TxEventUnmarshaler interface {
	UnmarshalTx(event string, data any) error
}

type TxEventParser struct {
	contract *blockchain.Contract
	log      *types.Log
}

func (t *TxEventParser) UnmarshalTx(name string, v any) error {
	if err := t.contract.ParseTxLog(name, t.log, v); err != nil {
		return err
	}
	if setter, ok := v.(CanSetTxHash); ok {
		setter.SetTxHash(t.log.TxHash)
	}
	return t.contract.ParseTxLog(name, t.log, v)
}
