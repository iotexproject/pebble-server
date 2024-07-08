package event

import (
	"context"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
)

func init() {
	f := func() Event { return &PebbleConfirm{} }
	e := f()
	registry(e.Topic(), f)
}

type PebbleConfirm struct {
	Imei    string
	Owner   common.Address
	Device  common.Address
	Channel uint32
}

func (e *PebbleConfirm) Source() SourceType { return SOURCE_TYPE__BLOCKCHAIN }

func (e *PebbleConfirm) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *PebbleConfirm) ContractID() string { return enums.CONTRACT__PEBBLE_DEVICE }

func (e *PebbleConfirm) EventName() string { return "Confirm" }

func (e *PebbleConfirm) Unmarshal(v any) error {
	return v.(TxEventUnmarshaler).UnmarshalTx(e.EventName(), e)
}

func (e *PebbleConfirm) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	if !contexts.CheckDeviceWhiteListFromContext(ctx, e.Imei) {
		return errors.Errorf("imei %s not in whitelist", e.Imei)
	}

	dev := &models.Device{
		ID:             e.Imei,
		Owner:          e.Owner.String(),
		Address:        e.Device.String(),
		Status:         models.CONFIRM,
		Proposer:       e.Owner.String(),
		OperationTimes: models.NewOperationTimes(),
	}
	_, err = UpsertOnConflict(ctx, dev, "id", "owner", "proposer", "status", "updated_at")
	return errors.Wrapf(err, "failed to upsert device: %s", dev.ID)
}
