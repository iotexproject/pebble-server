package event

import (
	"context"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/ethutil/address"
)

func init() {
	e := &AccountUpdated{}
	registry(e.Topic(), func() Event { return &AccountUpdated{} })
}

type AccountUpdated struct {
	owner  address.Address
	name   string
	avatar string
}

func (e *AccountUpdated) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *AccountUpdated) Topic() string {
	return "Updated(address owner, string name, string avatar)"
}

func (e *AccountUpdated) Unmarshal(data []byte) error {
	// unmarshal event log
	return nil
}

func (e *AccountUpdated) Handle(ctx context.Context) error {
	// upsert account
	return nil
}
