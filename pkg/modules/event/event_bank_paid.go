package event

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/xoctopus/x/misc/must"

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

	h common.Hash
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

func (e *BankPaid) Unmarshal(v any) error {
	log, ok := v.(*types.Log)
	must.BeTrueWrap(ok, "expect *types.Log to unmarshal `%t`, but got `%t`", e, v)
	e.h = log.TxHash
	return nil
}

func (e *BankPaid) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	br := &models.BankRecord{
		ID:        e.h.String() + "-" + e.Timestamp.String(),
		To:        e.To.String(),
		Amount:    e.Amount.String(),
		Timestamp: time.Now().Unix(),
		Type:      2,
	}

	err = UpsertOnConflictDoNothing(ctx, br, []string{"id"}, []*Assigner{
		{"id", br.ID},
		{"from", ""},
		{"to", br.To},
		{"amount", br.Amount},
		{"timestamp", br.Timestamp},
		{"type", br.Type},
		{"updated_at", time.Now()},
		{"created_at", time.Now()},
	}...)
	if err != nil {
		return
	}

	b := &models.Bank{
		Address: e.To.String(),
		Balance: e.Balance.String(),
	}
	return UpsertOnConflictUpdateOthers(ctx, b, []string{"address"}, []*Assigner{
		{"address", b.Address},
		{"balance", b.Balance},
		{"updated_at", time.Now()},
		{"created_at", time.Now()},
	}...)
}
