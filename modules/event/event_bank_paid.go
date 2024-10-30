package event

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

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

func (e *BankPaid) Source() enums.EventSourceType {
	return enums.EVENT_SOURCE_TYPE__BLOCKCHAIN
}

func (e *BankPaid) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *BankPaid) ContractID() string { return enums.CONTRACT__PEBBLE_BANK }

func (e *BankPaid) EventName() string { return "Paid" }

func (e *BankPaid) Unmarshal(v any) error {
	return v.(TxEventUnmarshaler).UnmarshalTx(e.EventName(), e)
}

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
		return errors.Wrapf(err, "failed to upsert bank_record: %s", br.ID)
	}

	_, err = UpsertOnConflict(ctx, b, "address", "balance", "updated_at")
	return errors.Wrapf(err, "failed to upsert bank: %s", b.Address)
}
