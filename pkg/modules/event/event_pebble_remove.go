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
	f := func() Event { return &PebbleRemove{} }
	e := f()
	registry(e.Topic(), f)
}

type PebbleRemove struct {
	IMEI
	Owner common.Address
}

func (e *PebbleRemove) Source() enums.EventSourceType {
	return enums.EVENT_SOURCE_TYPE__BLOCKCHAIN
}

func (e *PebbleRemove) Topic() string {
	return strings.Join([]string{
		"TOPIC", e.ContractID(), strings.ToUpper(e.EventName()),
	}, "__")
}

func (e *PebbleRemove) ContractID() string { return enums.CONTRACT__PEBBLE_DEVICE }

func (e *PebbleRemove) EventName() string { return "Remove" }

func (e *PebbleRemove) Unmarshal(v any) error {
	return v.(TxEventUnmarshaler).UnmarshalTx(e.EventName(), e)
}

func (e *PebbleRemove) Handle(ctx context.Context) (err error) {
	defer func() { err = WrapHandleError(err, e) }()

	if !contexts.CheckDeviceWhiteListFromContext(ctx, e.Imei) {
		return errors.Errorf("imei %s not in whitelist", e.Imei)
	}

	dev := &models.Device{ID: e.Imei}
	if err = FetchByPrimary(ctx, dev); err != nil {
		return errors.Wrapf(err, "failed to fetch device: %s", dev.ID)
	}
	if dev.Owner != e.Owner.String() {
		return errors.Errorf(
			"without device perimission: %s %s %s",
			e.Imei, dev.Owner, e.Owner.String(),
		)
	}
	dev.Owner = ""
	if dev.Status == models.CONFIRM {
		dev.Status = models.CREATED
	}
	if dev.Proposer == e.Owner.String() {
		dev.Proposer = ""
	}
	_, err = UpsertOnConflict(ctx, dev, "id", "owner", "status", "proposer")
	return errors.Wrapf(err, "failed to upsert device: %s", dev.ID)
}
