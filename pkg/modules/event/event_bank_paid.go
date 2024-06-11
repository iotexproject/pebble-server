package event

import (
	"context"
	"math/big"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/ethutil/address"
)

func init() {
	e := &BankPaid{}
	registry(e.Topic(), func() Event { return &BankPaid{} })
}

type BankPaid struct {
	tx      string
	from    address.Address
	to      address.Address
	amount  *big.Int
	ts      *big.Int
	balance *big.Int
}

func (e *BankPaid) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *BankPaid) Topic() string {
	return "Paid(address indexed from, address indexed to, uint256 amount, uint256 timestamp, uint256 balance)"
}

func (e *BankPaid) Unmarshal(data []byte) error {
	// unmarshal event log
	return nil
}

func (e *BankPaid) Handle(ctx context.Context) error {
	// create bank record
	// upsert bank
	// type 2
	// id = tx+ts.string()
	return nil
}
