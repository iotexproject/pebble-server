package event

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
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
}

func (e *BankWithdraw) Source() SourceType { return SOURCE_TYPE__BLOCKCHAIN }

func (e *BankWithdraw) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *BankWithdraw) ContractID() string { return enums.CONTRACT__PEBBLE_BANK }

func (e *BankWithdraw) EventName() string { return "Withdraw" }

func (e *BankWithdraw) Data() any { return e }

func (e *BankWithdraw) Unmarshal(any) error { return nil }

func (e *BankWithdraw) Handle(ctx context.Context) error {
	// create bank record
	// upsert bank
	// type 1
	return nil
}
