package event

import (
	"context"
	"math/big"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/ethutil/address"
)

func init() {
	e := &BankWithdraw{}
	registry(e.Topic(), func() Event { return &BankWithdraw{} })
}

type BankWithdraw struct {
	tx      string
	from    address.Address
	to      address.Address
	amount  *big.Int
	balance *big.Int
}

func (e *BankWithdraw) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *BankWithdraw) Topic() string {
	return "Withdraw(address indexed from, address indexed to, uint256 amount, uint256 balance)"
}

func (e *BankWithdraw) Unmarshal(data any) error {
	// unmarshal event log
	return nil
}

func (e *BankWithdraw) Handle(ctx context.Context) error {
	// create bank record
	// upsert bank
	// type 1
	return nil
}
