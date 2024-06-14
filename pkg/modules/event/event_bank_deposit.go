package event

import (
	"context"
	"math/big"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/ethutil/address"
)

func init() {
	e := &BankDeposit{}
	registry(e.Topic(), func() Event { return &BankDeposit{} })
}

type BankDeposit struct {
	tx      string
	to      address.Address
	amount  *big.Int
	balance *big.Int
}

func (e *BankDeposit) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *BankDeposit) Topic() string {
	return "Deposit(address indexed to, uint256 amount, uint256 balance)"
}

func (e *BankDeposit) Unmarshal(data any) error {
	// unmarshal event log
	return nil
}

func (e *BankDeposit) Handle(ctx context.Context) error {
	// create bank record
	// upsert bank
	// type 0
	return nil
}
