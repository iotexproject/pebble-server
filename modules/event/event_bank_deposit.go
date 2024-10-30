package event

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/iotexproject/pebble-server/enums"
	"github.com/iotexproject/pebble-server/models"
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
	TxHash
}

func (e *BankDeposit) Source() enums.EventSourceType {
	return enums.EVENT_SOURCE_TYPE__BLOCKCHAIN
}

func (e *BankDeposit) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *BankDeposit) ContractID() string { return enums.CONTRACT__PEBBLE_BANK }

func (e *BankDeposit) EventName() string { return "Deposit" }

func (e *BankDeposit) Unmarshal(v any) error {
	return v.(TxEventUnmarshaler).UnmarshalTx(e.EventName(), e)
}

func (e *BankDeposit) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	br := &models.BankRecord{
		ID:             e.hash.String(),
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
		return errors.Wrapf(err, "failed to upsert bank_record: %s", br.ID)
	}

	_, err = UpsertOnConflict(ctx, b, "address", "balance", "updated_at")
	return errors.Wrapf(err, "failed to upsert bank: %s", b.Address)
}
