package abi

import (
	"fmt"
	"math/big"
	"testing"
)

func TestEncode(t *testing.T) {
	v := big.NewInt(12343334)
	n, err := encodeNum(v)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%x", n)
}
