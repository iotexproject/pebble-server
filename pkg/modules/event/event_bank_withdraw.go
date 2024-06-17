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
		return &BankWithdraw{
			contractID: enums.CONTRACT__PEBBLE_BANK,
		}
	}
	e := f()
	registry(e.Topic(), f)
}

type BankWithdraw struct {
	From    common.Address
	To      common.Address
	Amount  *big.Int
	Balance *big.Int

	contractID string
}

func (e *BankWithdraw) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *BankWithdraw) Topic() string {
	return network.Topic(e.contractID) + "__" + strings.ToUpper(e.EventName())
}

func (e *BankWithdraw) ContractID() string {
	return network.ContractID(e.contractID)
}

func (e *BankWithdraw) EventName() string {
	return "Withdraw"
}

func (e *BankWithdraw) SubscriberID() string {
	return network.SubscriberID(e.contractID)
}

func (e *BankWithdraw) Data() any { return e }

func (e *BankWithdraw) Unmarshal(any) error { return nil }

func (e *BankWithdraw) Handle(ctx context.Context) error {
	// create bank record
	// upsert bank
	// type 1
	return nil
}
