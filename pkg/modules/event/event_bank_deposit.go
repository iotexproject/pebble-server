package event

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
)

func init() {
	f := func() Event {
		return &BankDeposit{
			contractID: enums.CONTRACT__PEBBLE_BANK,
		}
	}
	e := f()
	registry(e.Topic(), f)
}

type BankDeposit struct {
	To      common.Address
	Amount  *big.Int
	Balance *big.Int

	contractID string
}

func (e *BankDeposit) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *BankDeposit) Topic() string {
	return network.Topic(e.contractID) + "__" + strings.ToUpper(e.EventName())
}

func (e *BankDeposit) ContractID() string {
	return network.ContractID(e.contractID)
}

func (e *BankDeposit) EventName() string {
	return "Deposit"
}

func (e *BankDeposit) SubscriberID() string {
	return network.SubscriberID(e.contractID)
}

func (e *BankDeposit) Data() any { return e }

func (e *BankDeposit) Unmarshal(any) error { return nil }

func (e *BankDeposit) Handle(ctx context.Context) error {
	// create bank record
	// upsert bank
	// type 0
	return nil
}
