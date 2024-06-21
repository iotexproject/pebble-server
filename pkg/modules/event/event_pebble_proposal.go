package event

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
)

func init() {
	f := func() Event { return &PebbleProposal{} }
	e := f()
	registry(e.Topic(), f)
}

type PebbleProposal struct {
	Imei   string
	Owner  common.Address
	Device common.Address
	Name   string
	Avatar string
}

func (e *PebbleProposal) Source() SourceType { return SOURCE_TYPE__BLOCKCHAIN }

func (e *PebbleProposal) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *PebbleProposal) ContractID() string { return enums.CONTRACT__PEBBLE_DEVICE }

func (e *PebbleProposal) EventName() string { return "Proposal" }

func (e *PebbleProposal) Unmarshal(v any) error {
	return v.(TxEventUnmarshaler).UnmarshalTx(e.EventName(), e)
}

func (e *PebbleProposal) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	dev := &models.Device{
		ID:             e.Imei,
		Name:           e.Name,
		Address:        e.Device.String(),
		Avatar:         e.Avatar,
		Status:         models.PROPOSAL,
		Proposer:       e.Owner.String(),
		OperationTimes: models.NewOperationTimes(),
	}
	_, err = UpsertOnConflict(ctx, dev, "id", "name", "avatar", "status", "proposer", "updated_at")
	return errors.Wrapf(err, "failed to upsert device: %s", dev.ID)
}
