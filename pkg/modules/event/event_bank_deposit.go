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
	f := func() Event { return &BankDeposit{} }
	e := f()
	registry(e.Topic(), f)
}

type BankDeposit struct {
	To      common.Address
	Amount  *big.Int
	Balance *big.Int

	h common.Hash
}

func (e *BankDeposit) Source() SourceType { return SOURCE_TYPE__BLOCKCHAIN }

func (e *BankDeposit) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *BankDeposit) ContractID() string { return enums.CONTRACT__PEBBLE_BANK }

func (e *BankDeposit) EventName() string { return "Deposit" }

func (e *BankDeposit) Data() any { return e }

func (e *BankDeposit) Unmarshal(any) error { return nil }

func (e *BankDeposit) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	br := &models.BankRecord{
		ID:             e.h.String(),
		From:           "",
		To:             e.To.String(),
		Amount:         e.Amount.String(),
		Timestamp:      time.Now().Unix(),
		Type:           models.BankRecodeDeposit,
		OperationTimes: models.NewOperationTimes(),
	}
	b := &models.Bank{
		Address:        e.To.String(),
		Balance:        e.Balance.String(),
		OperationTimes: models.NewOperationTimes(),
	}

	_, err = UpsertOnConflict(ctx, br, "id")
	if err != nil {
		return err
	}

	_, err = UpsertOnConflict(ctx, b, "address", "balance", "updated_at")
	return err
}
