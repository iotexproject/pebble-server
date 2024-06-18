package event

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
)

func init() {
	f := func() Event { return &AccountUpdated{} }
	e := f()
	registry(e.Topic(), f)
}

type AccountUpdated struct {
	Owner  common.Address
	Name   string
	Avatar string
}

func (e *AccountUpdated) Source() SourceType { return SOURCE_TYPE__BLOCKCHAIN }

func (e *AccountUpdated) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *AccountUpdated) ContractID() string { return enums.CONTRACT__PEBBLE_ACCOUNT }

func (e *AccountUpdated) EventName() string { return "Updated" }

func (e *AccountUpdated) Unmarshal(v any) error {
	return v.(TxEventUnmarshaler).UnmarshalTx(e.EventName(), e)
}

func (e *AccountUpdated) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	m := &models.Account{
		ID:             e.Owner.String(),
		Name:           e.Name,
		Avatar:         e.Avatar,
		OperationTimes: models.NewOperationTimes(),
	}
	_, err = UpsertOnConflict(ctx, m, "id", "name", "avatar", "updated_at")
	return err
}
