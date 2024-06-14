package event

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/xoctopus/x/misc/must"
	"gorm.io/gorm/clause"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
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

func (e *AccountUpdated) Unmarshal(v any) error {
	_, ok := v.(types.Log)
	must.BeTrueWrap(ok, "assertion unmarshal with types.Log")

	return nil
}

func (e *AccountUpdated) Handle(ctx context.Context) error {
	db := must.BeTrueV(contexts.DatabaseFromContext(ctx))

	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "avatar"}),
	}).Create(&models.Account{
		ID:     e.owner.String(),
		Name:   e.name,
		Avatar: e.avatar,
	}).Error
	return err
}
