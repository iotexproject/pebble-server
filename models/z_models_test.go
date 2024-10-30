package models_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm/schema"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
)

func TestTableNames(t *testing.T) {
	r := require.New(t)

	for _, v := range []any{
		&models.Account{},
		&models.App{},
		&models.AppV2{},
		&models.Bank{},
		&models.BankRecord{},
		&models.Device{},
		&models.DeviceRecord{},
	} {
		m, ok := v.(schema.Tabler)
		r.True(ok)
		t.Log(reflect.TypeOf(m).String(), m.TableName())
	}
}
