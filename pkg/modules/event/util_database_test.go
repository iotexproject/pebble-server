package event_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xoctopus/x/contextx"
	"gorm.io/gorm"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/database"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/event"
)

func TestDatabaseOperations(t *testing.T) {
	t.Skip("need pg dependencies")

	r := require.New(t)

	d := &database.Postgres{}
	d.SetDefault()
	d.Endpoint.Username = "test"
	d.Endpoint.Password = "passwd"
	r.NoError(d.Init())

	ctx := contextx.WithContextCompose(
		contexts.WithDatabaseContext(d),
	)(context.Background())

	t.Run("UpsertOnConflict", func(t *testing.T) {
		m := &models.Account{
			ID:     "111",
			Name:   "name",
			Avatar: "avatar",
		}

		err := event.DeleteByPrimary(ctx, m, "111")

		_, err = event.UpsertOnConflict(ctx, m, "id")
		r.NoError(err)

		m.Avatar = "avatar2"
		v, err := event.UpsertOnConflict(ctx, m, "id")
		r.NoError(err)
		_, ok := v.(*models.Account)
		r.True(ok)
		err = event.FetchByPrimary(ctx, m)
		r.NoError(err)
		r.Equal(m.Avatar, "avatar")

		m.Avatar = "avatar3"
		v, err = event.UpsertOnConflict(ctx, m, "id", "avatar")
		_, ok = v.(*models.Account)
		r.True(ok)
		err = event.FetchByPrimary(ctx, m)
		r.NoError(err)
		r.Equal(m.Avatar, "avatar3")
	})

	t.Run("DeleteByPrimary", func(t *testing.T) {
		m := &models.Account{ID: "111"}
		err := event.DeleteByPrimary(ctx, m, "111")
		r.NoError(err)

		err = event.FetchByPrimary(ctx, m)
		r.ErrorIs(err, gorm.ErrRecordNotFound)
	})

	t.Run("UpdateByPrimary", func(t *testing.T) {
		m := &models.Account{ID: "111"}
		err := event.UpdateByPrimary(ctx, m, map[string]any{"avatar": "avatar4"})
		r.ErrorIs(err, gorm.ErrRecordNotFound)

		_, err = event.UpsertOnConflict(ctx, m, "id")
		r.NoError(err)

		err = event.UpdateByPrimary(ctx, m, map[string]any{"avatar": "avatar4"})
		r.NoError(err)
		r.Equal(m.Name, "")
		r.Equal(m.Avatar, "avatar4")
	})
}
