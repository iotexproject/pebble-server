package event_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	. "github.com/machinefi/sprout-pebble-sequencer/pkg/modules/event"
)

func TestQueryBuilderUpdate(t *testing.T) {
	r := require.New(t)

	now := time.Now()
	q, vs := BuildUpdateQuery(
		"account",
		[]*Assigner{
			{"id", 1},
			{"avatar", "any_url"},
			{"updated_at", now},
			{"created_at", now},
		}...,
	)

	r.Equal(*q, "UPDATE account SET id=?,avatar=?,updated_at=?,created_at=?")
	r.Len(vs, 4)
	r.Equal(vs[0], 1)
	r.Equal(vs[1], "any_url")
	r.Equal(vs[2], now)
	r.Equal(vs[3], now)

	q, vs = BuildUpdateQuery(
		"account",
		[]*Assigner{
			{"id", 1},
			{"avatar", "any_url"},
			{"updated_at", now},
			{"created_at", nil},
		}...,
	)
	r.Equal(*q, "UPDATE account SET id=?,avatar=?,updated_at=?")
	r.Len(vs, 3)
	r.Equal(vs[0], 1)
	r.Equal(vs[1], "any_url")
	r.Equal(vs[2], now)

	q, vs = BuildUpdateQuery("account")
	r.Nil(q)
	r.Nil(vs)
}

func TestQueryBuilderUpsert(t *testing.T) {
	r := require.New(t)

	now := time.Now()

	q, vs := BuildUpsertOnConflictUpdateOthersQuery(
		"account", []string{"id", "name"},
		[]*Assigner{
			{"id", 10},
			{"name", "fang"},
			{"org", "iotx"},
			{"date", now},
		}...,
	)
	r.Equal(*q, `INSERT INTO account (id,name,org,date) VALUES (?,?,?,?) ON CONFLICT (id,name) DO UPDATE SET date=?,org=?`)
	r.Len(vs, 4+4-2)
	r.Equal(vs[0], 10)
	r.Equal(vs[1], "fang")
	r.Equal(vs[2], "iotx")
	r.Equal(vs[3], now)
	r.Equal(vs[4], "iotx")
	r.Equal(vs[5], now)
}
