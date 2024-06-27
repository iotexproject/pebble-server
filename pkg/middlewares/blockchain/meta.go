package blockchain

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

func (mm *Meta) String() string {
	return fmt.Sprintf("%d_%s_%s", mm.Network, mm.Contract, mm.Topic)
}

func (mm *Meta) Bytes() []byte {
	return mm.MetaID().Bytes()
}

func (mm *Meta) TxHashKey(h common.Hash) []byte {
	return []byte(fmt.Sprintf("LOG_%s_%s", mm, h))
}

func (mm *Meta) BlockKey(tx *types.Log) []byte {
	return []byte(fmt.Sprintf("BLK_%s_%d_%s", mm, tx.BlockNumber, tx.TxHash))
}

func (mm *Meta) BlockKeyPrefixLowerBound(blk uint64) []byte {
	return []byte(fmt.Sprintf("BLK_%s_%d_", mm, blk))
}

func (mm *Meta) BlockKeyPrefixUpperBound(blk uint64) []byte {
	return []byte(fmt.Sprintf("BLK_%s_%d`", mm, blk))
}

func (mm *Meta) WatcherFromKey(sub string) []byte {
	return []byte(fmt.Sprintf("SUBL_%s_%s", mm, sub))
}

func (mm *Meta) WatcherEndKey(sub string) []byte {
	return []byte(fmt.Sprintf("SUBH_%s_%s", mm, sub))
}

func (mm *Meta) RangeEndKey() []byte {
	return []byte(fmt.Sprintf("RNH_%s", mm))
}

func (mm *Meta) RangeFromKey() []byte {
	return []byte(fmt.Sprintf("RNL_%s", mm))
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
