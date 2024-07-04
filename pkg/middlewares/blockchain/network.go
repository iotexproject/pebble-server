package blockchain

import (
	"bytes"

	"github.com/pkg/errors"
)

type Network int

const (
	NETWORK_UNKNOWN       Network = 0
	NETWORK__IOTX_MAINNET Network = 4689
	NETWORK__IOTX_TESTNET Network = 4690
)

var InvalidNetwork = errors.New("invalid Network type")

func ParseNetworkFromString(s string) (Network, error) {
	switch s {
	default:
		return NETWORK_UNKNOWN, InvalidNetwork
	case "":
		return NETWORK_UNKNOWN, nil
	case "IOTX_MAINNET":
		return NETWORK__IOTX_MAINNET, nil
	case "IOTX_TESTNET":
		return NETWORK__IOTX_TESTNET, nil
	}
}

func (v Network) Int() int {
	return int(v)
}

func (v Network) String() string {
	switch v {
	default:
		return "UNKNOWN"
	case NETWORK_UNKNOWN:
		return ""
	case NETWORK__IOTX_MAINNET:
		return "IOTX_MAINNET"
	case NETWORK__IOTX_TESTNET:
		return "IOTX_TESTNET"
	}
}

func (v Network) MarshalText() ([]byte, error) {
	s := v.String()
	if s == "UNKNOWN" {
		return nil, InvalidNetwork
	}
	return []byte(s), nil
}

func (v *Network) UnmarshalText(data []byte) error {
	s := string(bytes.ToUpper(data))
	val, err := ParseNetworkFromString(s)
	if err != nil {
		return err
	}
	*v = val
	return nil
}
