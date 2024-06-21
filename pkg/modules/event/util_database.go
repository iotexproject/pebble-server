package event

import (
	"context"

	"github.com/xoctopus/x/misc/must"
	"gorm.io/gorm/clause"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
)

func UpsertOnConflict(ctx context.Context, m any, conflict string, updates ...string) (any, error) {
	db := must.BeTrueV(contexts.DatabaseFromContext(ctx))

	cond := clause.OnConflict{
		Columns: []clause.Column{{Name: conflict}},
	}
	if len(updates) == 0 {
		cond.DoNothing = true
	} else {
		cond.DoUpdates = clause.AssignmentColumns(updates)
	}
	tx := db.Clauses(cond).Create(m)
	if err := tx.Error; err != nil {
		return nil, err
	}
	return m, nil
}

func DeleteByPrimary(ctx context.Context, m any) error {
	db := must.BeTrueV(contexts.DatabaseFromContext(ctx))
	return db.Delete(m).Error
}

func UpdateByPrimary(ctx context.Context, m any, fields map[string]any) error {
	db := must.BeTrueV(contexts.DatabaseFromContext(ctx))
	if err := db.Model(m).Updates(fields).Error; err != nil {
		return err
	}
	return FetchByPrimary(ctx, m)
}

func FetchByPrimary(ctx context.Context, m any) error {
	db := must.BeTrueV(contexts.DatabaseFromContext(ctx))
	return db.First(m).Error
}
