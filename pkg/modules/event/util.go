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
	return false
}

var gByteOrder = binary.BigEndian
