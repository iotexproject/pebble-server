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
	f := func() Event { return &BankPaid{} }
	e := f()
	registry(e.Topic(), f)
}

type BankPaid struct {
	From      common.Address
	To        common.Address
	Amount    *big.Int
	Timestamp *big.Int
	Balance   *big.Int
	TxHash
}

func (e *BankPaid) Source() SourceType { return SOURCE_TYPE__BLOCKCHAIN }

func (e *BankPaid) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *BankPaid) ContractID() string { return enums.CONTRACT__PEBBLE_BANK }

func (e *BankPaid) EventName() string { return "Paid" }

func (e *BankPaid) Data() any { return e }

func (e *BankPaid) Unmarshal(any) error { return nil }

func (e *BankPaid) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	br := &models.BankRecord{
		ID:             e.hash.String() + "-" + e.Timestamp.String(),
		From:           e.From.String(),
		To:             e.To.String(),
		Amount:         e.Amount.String(),
		Timestamp:      time.Now().Unix(),
		Type:           models.BankRecodePaid,
		OperationTimes: models.NewOperationTimes(),
	}
	b := &models.Bank{
		Address:        e.From.String(),
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
