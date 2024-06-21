package event_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xoctopus/confx/confmws/confmqtt"
	"github.com/xoctopus/datatypex"
	"github.com/xoctopus/x/contextx"
	"github.com/xoctopus/x/misc/retry"
	"gorm.io/gorm"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/crypto"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/database"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/event"
)

func testctx() context.Context {
	d := &database.Postgres{}
	d.SetDefault()
	d.Endpoint.Username = "test"
	d.Endpoint.Password = "passwd"
	if err := d.Init(); err != nil {
		return nil
	}

	mq := &confmqtt.Broker{
		Server:        datatypex.Endpoint{},
		Retry:         retry.Retry{},
		Timeout:       0,
		Keepalive:     0,
		RetainPublish: false,
		QoS:           0,
		Cert:          nil,
	}
	mq.SetDefault()
	if err := mq.Init(); err != nil {
		return nil
	}
	sk := &crypto.EcdsaPrivateKey{
		Hex: "dbfe03b0406549232b8dccc04be8224fcc0afa300a33d4f335dcfdfead861c85",
	}
	if err := sk.Init(); err != nil {
		return nil
	}

	ctx := contextx.WithContextCompose(
		contexts.WithDatabaseContext(d),
		contexts.WithProjectIDContext(1),
		contexts.WithProjectVersionContext("v0.0.1"),
		contexts.WithMqttBrokerContext(mq),
		contexts.WithEcdsaPrivateKeyContext(sk),
	)(context.Background())

	if _, err := event.UpsertOnConflict(ctx, &models.Device{
		ID:                     "350916067070535",
		Name:                   "Kavoit 2",
		Owner:                  "0x4F3794b74F9d59C5773a7272295A6A3F0cCc972a",
		Address:                "0x7c6D376B125b43ea5ba5dd0F84E00a4f0c3ad63F",
		Avatar:                 "https://storageapi.fleek.co/uu-z-team-bucket/b7a5acf3-c513-4279-944b-72f4dbf17e8c",
		Status:                 2,
		Proposer:               "",
		Firmware:               "Riverrock 1.0.14",
		Config:                 "",
		TotalGas:               0,
		BulkUpload:             0,
		DataChannel:            8183,
		UploadPeriod:           300,
		BulkUploadSamplingCnt:  0,
		BulkUploadSamplingFreq: 0,
		Beep:                   100,
		RealFirmware:           "Riverrock 1.0.14",
		State:                  0,
		Type:                   0,
		Configurable:           false,
	}, "id"); err != nil {
		return nil
	}
	if _, err := event.UpsertOnConflict(ctx, &models.Device{
		ID:                     "351358815439952",
		Name:                   "Pebble1",
		Owner:                  "0x8bf170A0274AE906B88b2234eC95489b60dea57E",
		Address:                "0x0c9B334A4c8CF6070F057e97FB7612084565842E",
		Avatar:                 "https://storageapi.fleek.co/uu-z-team-bucket/b7a5acf3-c513-4279-944b-72f4dbf17e8c",
		Status:                 2,
		Proposer:               "",
		Firmware:               "",
		Config:                 "",
		TotalGas:               0,
		BulkUpload:             0,
		DataChannel:            8183,
		UploadPeriod:           300,
		BulkUploadSamplingCnt:  0,
		BulkUploadSamplingFreq: 0,
		Beep:                   1000,
		RealFirmware:           "Riverrock 1.0.12",
		State:                  0,
		Type:                   0,
		Configurable:           false,
	}, "id"); err != nil {
		return nil
	}
	return ctx
}

func TestDatabaseOperations(t *testing.T) {
	// t.Skip("need pg dependencies")

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
