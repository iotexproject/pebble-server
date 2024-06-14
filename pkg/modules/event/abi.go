package event

import (
	"bytes"
	_ "embed"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/xoctopus/x/misc/must"
)

var (
	//go:embed abis/account_updated.json
	accountUpdated []byte
	//go:embed abis/bank_deposit.json
	bankDeposit []byte
	//go:embed abis/bank_paid.json
	bankPaid []byte
	//go:embed abis/bank_withdraw.json
	bankWithdraw []byte
	//go:embed abis/firmware_removed.json
	firmwareRemoved []byte
	//go:embed abis/firmware_updated.json
	firmwareUpdated []byte
	//go:embed abis/pebble_config.json
	pebbleConfig []byte
	//go:embed abis/pebble_confirmed.json
	pebbleConfirmed []byte
	//go:embed abis/pebble_firmware.json
	pebbleFirmware []byte
	//go:embed abis/pebble_proposed.json
	pebbleProposed []byte
	//go:embed abis/pebble_removed.json
	pebbleRemoved []byte

	accountUpdatedParser  = must.NoErrorV(NewEventParser("Updated", must.NoErrorV(abi.JSON(bytes.NewReader(accountUpdated)))))
	bankDepositParser     = must.NoErrorV(NewEventParser("Deposit", must.NoErrorV(abi.JSON(bytes.NewReader(bankDeposit)))))
	bankPaidParser        = must.NoErrorV(NewEventParser("Paid", must.NoErrorV(abi.JSON(bytes.NewReader(bankPaid)))))
	bankWithdrawParser    = must.NoErrorV(NewEventParser("Withdraw", must.NoErrorV(abi.JSON(bytes.NewReader(bankWithdraw)))))
	firmwareRemovedParser = must.NoErrorV(NewEventParser("FirmwareRemoved", must.NoErrorV(abi.JSON(bytes.NewReader(firmwareRemoved)))))
	firmwareUpdatedParser = must.NoErrorV(NewEventParser("FirmwareUpdated", must.NoErrorV(abi.JSON(bytes.NewReader(firmwareUpdated)))))
	pebbleConfigParser    = must.NoErrorV(NewEventParser("Config", must.NoErrorV(abi.JSON(bytes.NewReader(pebbleConfig)))))
	pebbleConfirmedParser = must.NoErrorV(NewEventParser("Confirm", must.NoErrorV(abi.JSON(bytes.NewReader(pebbleConfirmed)))))
	pebbleFirmwareParser  = must.NoErrorV(NewEventParser("Firmware", must.NoErrorV(abi.JSON(bytes.NewReader(pebbleFirmware)))))
	pebbleProposedParser  = must.NoErrorV(NewEventParser("Proposal", must.NoErrorV(abi.JSON(bytes.NewReader(pebbleProposed)))))
	pebbleRemovedParser   = must.NoErrorV(NewEventParser("Remove", must.NoErrorV(abi.JSON(bytes.NewReader(pebbleRemoved)))))
)

func NewEventParser(name string, abi abi.ABI) (*EventParser, error) {
	event, ok := abi.Events[name]
	if !ok {
		return nil, errEventABINotFound
	}

	return &EventParser{
		name:  name,
		abi:   abi,
		event: event,
	}, nil
}

type EventParser struct {
	name  string
	abi   abi.ABI
	event abi.Event
}

func (p *EventParser) Parse(l *types.Log, result any) error {
	rv, ok := result.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(result)
	}
	if !rv.IsValid() {
		return errors.Errorf("expect valid result, but got nil")
	}

	rt := rv.Type()
	if rt.Kind() == reflect.Pointer {
		if rv.IsNil() && rv.CanSet() {
			rv.Set(reflect.New(rt.Elem()))
		}
		return p.Parse(l, rv.Elem())
	}

	if !rv.CanSet() {
		return errors.Errorf("expect result can be set")
	}

	if len(l.Topics) == 0 {
		return errNoEventSignature
	}
	if l.Topics[0] != p.abi.Events[p.name].ID {
		return errEventSignatureMismatch
	}
	if len(l.Data) > 0 {
		if err := p.abi.UnpackIntoInterface(nil, p.name, l.Data); err != nil {
			return err
		}
	}
	var indexed abi.Arguments
	for _, arg := range p.abi.Events[p.name].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return abi.ParseTopics(result, indexed, l.Topics[1:])
}

var (
	errEventABINotFound       = errors.New("event abi not found")
	errContractMismatch       = errors.New("contract mismatch")
	errNoEventSignature       = errors.New("no event signature")
	errEventSignatureMismatch = errors.New("event sig mismatch")
)
