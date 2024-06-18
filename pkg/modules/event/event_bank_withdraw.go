package event

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
)

func init() {
	f := func() Event { return &BankWithdraw{} }
	e := f()
	registry(e.Topic(), f)
}

type BankWithdraw struct {
	From    common.Address
	To      common.Address
	Amount  *big.Int
	Balance *big.Int
	TxHash
}

func (e *BankWithdraw) Source() SourceType { return SOURCE_TYPE__BLOCKCHAIN }

func (e *BankWithdraw) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *BankWithdraw) ContractID() string { return enums.CONTRACT__PEBBLE_BANK }

func (e *BankWithdraw) EventName() string { return "Withdraw" }

func (e *BankWithdraw) Unmarshal(v any) error {
	return v.(TxEventUnmarshaler).UnmarshalTx(e.EventName(), e)
}

func (e *BankWithdraw) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	br := &models.BankRecord{
		ID:        e.hash.String(),
		From:      e.From.String(),
		To:        e.To.String(),
		Amount:    e.Amount.String(),
		Timestamp: time.Now().Unix(),
		Type:      models.BankRecodeWithdraw,
	}

	b := &models.Bank{
		Address: e.From.String(),
		Balance: e.Balance.String(),
	}

	_, err = UpsertOnConflict(ctx, br, "id")
	if err != nil {
		return err
	}

	_, err = UpsertOnConflict(ctx, b, "address", "balance", "updated_at")
	return err
}
