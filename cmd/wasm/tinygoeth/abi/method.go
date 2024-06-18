package abi

import (
	"fmt"
	"strings"

	"github.com/machinefi/sprout-pebble-sequencer/cmd/wasm/tinygoeth/common"
)

// currently tinygo does not support compile array, tuple, function, fixed bytes
type Method struct {
	Name   string
	Sig    string
	Inputs []Type
}

// NewMethod creates a new method from a string signature
// signature format: "method_name(type,type,...)"
func NewMethod(sig string) (*Method, error) {
	m := &Method{
		Sig: sig,
	}
	var inputs []string
	for i, c := range sig {
		if c == '(' {
			m.Name = sig[:i]
			sig = sig[i+1:]
			break
		}
	}
	sigLen := len(sig)
	inputs = strings.Split(sig[:sigLen-1], ",")
	for _, input := range inputs {
		t, err := NewType(input)
		if err != nil {
			return nil, err
		}
		m.Inputs = append(m.Inputs, t)
	}
	return m, nil
}

// ID returns the id of the method
func (m *Method) ID() []byte {
	return common.Keccak256([]byte(m.Sig))[:4]
}

// Pack packs the given arguments into a byte slice according to the method signature
func (m *Method) Pack(args ...interface{}) ([]byte, error) {
	if len(args) != len(m.Inputs) {
		return nil, fmt.Errorf("expected %d arguments, got %d", len(m.Inputs), len(args))
	}
	abiArgs := m.Inputs

	// variable input is the output appended at the end of packed
	// output. This is used for strings and bytes types input.
	var variableInput []byte

	// input offset is the bytes offset for packed output
	inputOffset := 0
	for _, abiArg := range abiArgs {
		inputOffset += getTypeSize(abiArg)
	}
	var ret []byte
	for i, a := range args {
		input := abiArgs[i]
		// pack the input
		packed, err := input.Encode(a)
		if err != nil {
			return nil, err
		}
		// check for dynamic types
		if isDynamicType(input) {
			// set the offset
			n, _ := encodeNum(inputOffset)
			ret = append(ret, n...)
			// calculate next offset
			inputOffset += len(packed)
			// append to variable input
			variableInput = append(variableInput, packed...)
		} else {
			// append the packed value to the input
			ret = append(ret, packed...)
		}
	}
	// append the variable input at the end of the packed input
	ret = append(ret, variableInput...)

	// Pack up the method ID too if not a constructor and return
	return append(m.ID(), ret...), nil
}

func (m *Method) String() string {
	var strs []string
	for _, t := range m.Inputs {
		strs = append(strs, t.String())
	}
	return fmt.Sprintf("%s(%s)", m.Name, strings.Join(strs, ","))
}
