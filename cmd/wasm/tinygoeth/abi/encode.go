package abi

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/machinefi/sprout-pebble-sequencer/cmd/wasm/tinygoeth/address"
)

// Encode encodes a value
func Encode(v interface{}, t Type) ([]byte, error) {
	return encode(v, t)
}
func encode(v interface{}, t Type) ([]byte, error) {
	switch t.T {
	// case SliceTy:
	// 	return encodeSliceAndArray(v, t)

	// case KindTuple:
	// 	return encodeTuple(v, t)

	case StringTy:
		return encodeString(v.(string))

	case BoolTy:
		return encodeBool(v.(bool))

	case AddressTy:
		return encodeAddress(v.(address.Address))

	case IntTy, UintTy:
		return encodeNum(v)

	case BytesTy:
		return encodeBytes(v.([]byte))

	// case KindFixedBytes, KindFunction:
	// 	return encodeFixedBytes(v)

	default:
		return nil, fmt.Errorf("encoding not available for type '%v'", t.T)
	}
}

var (
	zero = big.NewInt(0)
	one  = big.NewInt(1)
)

func encodeString(v string) ([]byte, error) {

	return packBytesSlice([]byte(v), len(v))
}

func encodeBool(v bool) ([]byte, error) {
	if v {
		return leftPad(one.Bytes(), 32), nil
	}
	return leftPad(zero.Bytes(), 32), nil
}

func encodeBytes(v []byte) ([]byte, error) {
	return packBytesSlice(v, len(v))
}

func packBytesSlice(buf []byte, l int) ([]byte, error) {
	len, err := encodeNum(l)
	if err != nil {
		return nil, err
	}
	return append(len, rightPad(buf, (l+31)/32*32)...), nil
}

func encodeAddress(v address.Address) ([]byte, error) {
	return leftPad(v.Bytes(), 32), nil
}

func encodeNum(v interface{}) ([]byte, error) {
	switch v.(type) {
	case uint:
		return toU256(new(big.Int).SetUint64(uint64(v.(uint)))), nil
	case uint8:
		return toU256(new(big.Int).SetUint64(uint64(v.(uint8)))), nil
	case uint16:
		return toU256(new(big.Int).SetUint64(uint64(v.(uint16)))), nil
	case uint32:
		return toU256(new(big.Int).SetUint64(uint64(v.(uint32)))), nil
	case uint64:
		return toU256(new(big.Int).SetUint64(v.(uint64))), nil
	case int:
		return toU256(new(big.Int).SetInt64(int64(v.(int)))), nil
	case int8:
		return toU256(new(big.Int).SetInt64(int64(v.(int8)))), nil
	case int16:
		return toU256(new(big.Int).SetInt64(int64(v.(int16)))), nil
	case int32:
		return toU256(new(big.Int).SetInt64(int64(v.(int32)))), nil
	case int64:
		return toU256(new(big.Int).SetInt64(v.(int64))), nil
	case *big.Int:
		return toU256(v.(*big.Int)), nil
	default:
		return nil, encodeErr(v, "number")
	}
}
func encodeErr(v interface{}, t string) error {
	return fmt.Errorf("failed to encode %v as %s", v, t)
}

var (
	tt256   = new(big.Int).Lsh(big.NewInt(1), 256)   // 2 ** 256
	tt256m1 = new(big.Int).Sub(tt256, big.NewInt(1)) // 2 ** 256 - 1
)

// U256 converts a big Int into a 256bit EVM number.
func toU256(n *big.Int) []byte {
	b := new(big.Int)
	b = b.Set(n)

	if b.Sign() < 0 || b.BitLen() > 256 {
		b.And(b, tt256m1)
	}

	return leftPad(b.Bytes(), 32)
}

func padBytes(b []byte, size int, left bool) []byte {
	l := len(b)
	if l == size {
		return b
	}
	if l > size {
		return b[l-size:]
	}
	tmp := make([]byte, size)
	if left {
		copy(tmp[size-l:], b)
	} else {
		copy(tmp, b)
	}
	return tmp
}

func leftPad(b []byte, size int) []byte {
	return padBytes(b, size, true)
}

func rightPad(b []byte, size int) []byte {
	return padBytes(b, size, false)
}

func encodeHex(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}

func decodeHex(str string) ([]byte, error) {
	if strings.HasPrefix(str, "0x") {
		str = str[2:]
	}
	buf, err := hex.DecodeString(str)
	if err != nil {
		return nil, fmt.Errorf("could not decode hex: %v", err)
	}
	return buf, nil
}
