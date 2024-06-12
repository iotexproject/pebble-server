package abi

import (
	"fmt"
	"strings"

	"github.com/machinefi/sprout-pebble-sequencer/cmd/wasm/tinygoeth/common"
)

type Event struct {
	// RawName is the raw event name parsed from ABI.
	RawName string
	Inputs  Arguments
	str     string
	// Sig contains the string signature according to the ABI spec.
	// e.g.	 event foo(uint32 a, int b) = "foo(uint32,int256)"
	// Please note that "int" is substitute for its canonical representation "int256"
	Sig string
	// ID returns the canonical representation of the event's signature used by the
	// abi definition to identify event names and types.
	ID common.Hash
}

// NewEvent creates a new event from a string signature
// signature format: "FirmwareUpdated(string name, string version, string uri, string avatar,address indexed from)"
func NewEvent(event string) (*Event, error) {
	e := &Event{
		str: event,
	}
	var inputs []string
	for i, c := range event {
		if c == '(' {
			e.RawName = event[:i]
			event = event[i+1:]
			break
		}
	}
	eventLen := len(event)
	inputs = strings.Split(event[:eventLen-1], ",")

	for _, input := range inputs {
		arr := strings.Split(input, " ")
		arr = removeEmpty(arr)
		if len(arr) != 2 && len(arr) != 3 {
			return nil, fmt.Errorf("invalid event input: %s", input)
		}
		eType, eIndexed, eName := parseEventInput(arr)
		t, err := NewType(eType)
		if err != nil {
			return nil, err
		}

		e.Inputs = append(e.Inputs, Argument{
			Name:    eName,
			Type:    t,
			Indexed: eIndexed,
		})
	}
	e.Sig = e.RawName + "(" + e.Inputs.String() + ")"
	e.ID = common.BytesToHash(common.Keccak256([]byte(e.Sig)))
	return e, nil
}

func (e *Event) Unpack(data []byte) ([]interface{}, error) {
	return e.Inputs.Unpack(data)
}

func (e *Event) String() string {
	return e.str
}

func removeEmpty(arr []string) []string {
	var ret []string
	for _, s := range arr {
		if s != "" {
			ret = append(ret, s)
		}
	}
	return ret
}
func parseEventInput(arr []string) (string, bool, string) {
	if len(arr) == 3 {
		return arr[0], arr[1] == "indexed", arr[2]
	}
	return arr[0], false, arr[1]
}
