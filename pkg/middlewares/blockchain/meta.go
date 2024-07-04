package blockchain

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

func NewMeta(network Network, contracts ...*Contract) *Meta {
	m := &Meta{
		Network:   network,
		Topics:    [][]common.Hash{{}},
		contracts: make(map[string]*Contract),
	}

	buf := bytes.NewBuffer(nil)
	_ = binary.Write(buf, gByteOrder, uint32(network))
	for _, c := range contracts {
		if c.Network != network {
			continue
		}
		m.contracts[c.Address.String()] = c
		m.Contracts = append(m.Contracts, c.Address)
		buf.Write(c.Address.Bytes())
		for _, e := range c.Events {
			m.Topics[0] = append(m.Topics[0], e.event.ID)
			buf.Write(e.event.ID.Bytes())
		}
	}

	m.ID = hex.EncodeToString(crypto.Keccak256(buf.Bytes()))
	return m
}

type Meta struct {
	ID        string           `json:"id"`
	Network   Network          `json:"network"`
	Contracts []common.Address `json:"contracts"`
	Topics    [][]common.Hash  `json:"topics"`

	contracts map[string]*Contract
}

func (m *Meta) String() string {
	return m.ID
}

func (m *Meta) MonitorRangeFromKey() []byte {
	return []byte(fmt.Sprintf("RNL_%s", m.ID))
}

func (m *Meta) MonitorRangeEndKey() []byte {
	return []byte(fmt.Sprintf("RNH_%s", m.ID))
}

func (m *Meta) WatcherRangeFromKey(sub string) []byte {
	return []byte(fmt.Sprintf("SUBL_%s_%s", m.ID, sub))
}

func (m *Meta) WatcherRangeEndKey(sub string) []byte {
	return []byte(fmt.Sprintf("SUBH_%s_%s", m.ID, sub))
}

func (m *Meta) BlockKeyPrefixLowerBound(blk uint64) []byte {
	return []byte(fmt.Sprintf("LOG_%d_%d_", m.Network, blk))
}

func (m *Meta) BlockKeyPrefixUpperBound(blk uint64) []byte {
	return append([]byte(fmt.Sprintf("LOG_%d_%d_", m.Network, blk)), 0xFF)
}

func (m *Meta) LogKey(l *types.Log) []byte {
	return []byte(fmt.Sprintf("LOG_%d_%d_%d_%s_%s_%s",
		m.Network, l.BlockNumber, l.TxIndex, l.Address, l.Topics[0], l.TxHash))
}

func (m *Meta) ParseLogKey(key string) (*Contract, string, error) {
	parts := strings.Split(key, "_")
	if len(parts) != 7 {
		return nil, "", errors.New("invalid log key parts")
	}

	i := 0
	if parts[i] != "LOG" {
		return nil, "", errors.New("invalid log key prefix")
	}
	i++
	network, err := strconv.ParseUint(parts[i], 10, 32)
	if err != nil || int(network) != m.Network.Int() {
		return nil, "", errors.New("invalid log key network")
	}
	i++
	if _, err = strconv.ParseUint(parts[i], 10, 64); err != nil {
		return nil, "", errors.Wrap(err, "invalid log key block number")
	}
	i++
	if _, err = strconv.ParseUint(parts[i], 10, 64); err != nil {
		return nil, "", errors.Wrap(err, "invalid log key tx index")
	}
	i++
	c, ok := m.contracts[parts[i]]
	if !ok {
		return nil, "", errors.New("contract not found in meta")
	}
	i++
	event := ""
	h := common.HexToHash(parts[i])
	for _, v := range c.events {
		if v.event.ID == h {
			event = v.event.Name
			break
		}
	}
	if event == "" {
		return nil, "", errors.New("event not found")
	}
	return c, event, nil
}
