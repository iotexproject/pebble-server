package event

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/xoctopus/x/misc/must"
	"gorm.io/gorm/clause"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
)

func init() {
	f := func() Event {
		return &AccountUpdated{
			contractID: enums.CONTRACT__PEBBLE_ACCOUNT,
		}
	}
	e := f()
	registry(e.Topic(), f)
}

type AccountUpdated struct {
	Owner  common.Address
	Name   string
	Avatar string

	contractID string
}

var _ EventHasBlockchainMeta = (*AccountUpdated)(nil)

func (e *AccountUpdated) Source() SourceType {
	return SourceTypeBlockchain
}

func (e *AccountUpdated) Topic() string {
	return network.Topic(e.contractID) + "__" + strings.ToUpper(e.EventName())
}

func (e *AccountUpdated) ContractID() string {
	return network.ContractID(e.contractID)
}

func (e *AccountUpdated) EventName() string {
	return "Updated"
}

func (e *AccountUpdated) SubscriberID() string {
	return network.SubscriberID(e.contractID)
}

func (e *AccountUpdated) Data() any { return e }

func (e *AccountUpdated) Unmarshal(any) error { return nil }

func (e *AccountUpdated) Handle(ctx context.Context) error {
	db := must.BeTrueV(contexts.DatabaseFromContext(ctx))

	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "avatar"}),
	}).Create(&models.Account{
		ID:     e.Owner.String(),
		Name:   e.Name,
		Avatar: e.Avatar,
	}).Error
	return err
}
