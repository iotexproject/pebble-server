package common

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"hash"

	"github.com/dustinxie/ecc"
	"github.com/machinefi/sprout-pebble-sequencer/cmd/wasm/tinygoeth/address"
	"golang.org/x/crypto/sha3"
)

// KeccakState wraps sha3.state. In addition to the usual hash methods, it also supports
// Read to get a variable amount of data from the hash state. Read is faster than Sum
// because it doesn't copy the internal state, but also modifies the internal state.
type KeccakState interface {
	hash.Hash
	Read([]byte) (int, error)
}

// NewKeccakState creates a new KeccakState
func NewKeccakState() KeccakState {
	return sha3.NewLegacyKeccak256().(KeccakState)
}

func Keccak256(data ...[]byte) []byte {
	b := make([]byte, 32)
	d := NewKeccakState()
	for _, b := range data {
		d.Write(b)
	}
	d.Read(b)
	return b
}

func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	curve := ecc.P256k1()
	if !curve.IsOnCurve(pub.X, pub.Y) {
		return nil
	}
	return elliptic.Marshal(curve, pub.X, pub.Y)
}

func PubkeyToAddress(p *ecdsa.PublicKey) address.Address {
	pubBytes := FromECDSAPub(p)
	if len(pubBytes) == 0 {
		return address.Address{}
	}
	return address.BytesToAddress(Keccak256(pubBytes[1:])[12:])
}

func RecoverPubkey(hash, sig []byte) (*ecdsa.PublicKey, error) {
	return ecc.RecoverPubkey("P-256k1", hash, sig)
}
