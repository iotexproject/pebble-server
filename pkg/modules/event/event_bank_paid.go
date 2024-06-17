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
		return &BankPaid{
			contractID: enums.CONTRACT__PEBBLE_BANK,
		}
	}
	e := f()
	registry(e.Topic(), f)
}

type BankPaid struct {
	From      common.Address
	To        common.Address
	Amount    *big.Int
	Timestamp *big.Int
	Balance   *big.Int

	contractID string
}

func (e *BankPaid) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *BankPaid) Topic() string {
	return network.Topic(e.contractID) + "__" + strings.ToUpper(e.EventName())
}

func (e *BankPaid) ContractID() string {
	return network.ContractID(e.contractID)
}

func (e *BankPaid) EventName() string {
	return "Paid"
}

func (e *BankPaid) SubscriberID() string {
	return network.SubscriberID(e.contractID)
}

func (e *BankPaid) Data() any { return e }

func (e *BankPaid) Unmarshal(any) error { return nil }

func (e *BankPaid) Handle(ctx context.Context) error {
	// create bank record
	// upsert bank
	// type 2
	// id = tx+ts.string()
	return nil
}
