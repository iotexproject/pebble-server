package blockchain

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
)

const MetaIDLength = 4 + common.AddressLength + common.HashLength

type MetaID [MetaIDLength]byte

func (mi MetaID) Bytes() []byte {
	return mi[:]
}

func (mi MetaID) String() string {
	return hex.EncodeToString(mi[:])
}

func ParseMonitorMeta(data []byte) (*Meta, error) {
	mm := &Meta{}
	if err := mm.UnmarshalText(data); err != nil {
		return nil, err
	}
	return mm, nil
}

type Meta struct {
	Network  Network
	Contract common.Address
	Topic    common.Hash
}

func (mm *Meta) MetaID() MetaID {
	buf := MetaID{}
	offset := 0
	gByteOrder.PutUint32(buf[offset:], uint32(mm.Network))
	offset += 4
	copy(buf[offset:], mm.Contract[:])
	offset += common.AddressLength
	copy(buf[offset:], mm.Topic[:])
	return buf
}

func (mm Meta) MarshalText() ([]byte, error) {
	return mm.MetaID().Bytes(), nil
}

func (mm *Meta) UnmarshalText(data []byte) error {
	if len(data) < MetaIDLength {
		return errors.Errorf("expect serilized data length greater than 68")
	}
	offset := 0
	mm.Network = Network(gByteOrder.Uint32(data[offset:]))
	offset += 4
	copy(mm.Contract[:], data[offset:offset+common.AddressLength])
	offset += common.AddressLength
	copy(mm.Topic[:], data[offset:])
	return nil
}
