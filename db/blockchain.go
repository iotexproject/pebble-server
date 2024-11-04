package db

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type scannedBlockNumber struct {
	gorm.Model
	Number uint64 `gorm:"not null"`
}

func (d *DB) ScannedBlockNumber() (uint64, error) {
	t := scannedBlockNumber{}
	if err := d.db.Where("id = ?", 1).First(&t).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, errors.Wrap(err, "failed to query scanned block number")
	}
	return t.Number, nil
}

func (d *DB) UpsertScannedBlockNumber(number uint64) error {
	t := scannedBlockNumber{
		Model: gorm.Model{
			ID: 1,
		},
		Number: number,
	}
	err := d.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"number"}),
	}).Create(&t).Error
	return errors.Wrap(err, "failed to upsert scanned block number")
}
