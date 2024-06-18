//go:build !tinygo

package abi_test

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
)

var (
	tt256   = new(big.Int).Lsh(big.NewInt(1), 256)   // 2 ** 256
	tt256m1 = new(big.Int).Sub(tt256, big.NewInt(1)) // 2 ** 256 - 1

	bigIntT = reflect.TypeOf(new(big.Int))
)

func packNum(offset int) []byte {
	n, _ := encodeNum(reflect.ValueOf(offset))
	return n
}

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
func encodeNum(v reflect.Value) ([]byte, error) {
	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return toU256(new(big.Int).SetUint64(v.Uint())), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return toU256(big.NewInt(v.Int())), nil

	case reflect.Ptr:
		if v.Type() != bigIntT {
			return nil, encodeErr(v.Elem(), "number")
		}
		return toU256(v.Interface().(*big.Int)), nil

	case reflect.Float64:
		return encodeNum(reflect.ValueOf(int64(v.Float())))

	case reflect.String:
		n, ok := new(big.Int).SetString(v.String(), 10)
		if !ok {
			n, ok = new(big.Int).SetString(v.String()[2:], 16)
			if !ok {
				return nil, encodeErr(v, "number")
			}
		}
		return encodeNum(reflect.ValueOf(n))

	default:
		return nil, encodeErr(v, "number")
	}
}

func encodeErr(v interface{}, t string) error {
	return fmt.Errorf("failed to encode %v as %s", v, t)
}

func TestXxx(t *testing.T) {
	v := big.NewInt(12343334)
	n, err := encodeNum(reflect.ValueOf(v))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%x", n)
}
