package event

import (
	"encoding/binary"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/ethutil/address"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/modules/ethutil/common"
)

func ValidateSignature(sum, sig []byte, addr address.Address) bool {
	for i := 0; i < 4; i++ {
		sig[64] = byte(i)
		pk, err := common.RecoverPubkey(sum, sig)
		if err != nil {
			continue
		}
		if common.PubkeyToAddress(pk) == addr {
			return true
		}
	}
	/*
		for i := 0; i < 4; i++ {
			sig[64] = byte(i)
			pk, err := ecc.RecoverPubkey("P-256k1", sum, sig)
			if err != nil {
				continue
			}
			if pk != nil && pk.X != nil && pk.Y != nil &&
				ecc.P256k1().IsOnCurve(pk.X, pk.Y) {
				raw := elliptic.Marshal(ecc.P256k1(), pk.X, pk.Y)
			}
		}

	*/
	return false
}

var gByteOrder = binary.BigEndian
