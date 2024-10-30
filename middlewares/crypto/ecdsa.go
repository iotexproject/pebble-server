package crypto

import (
	"crypto/ecdsa"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/crypto"
)

const MaskedPrivateKey = "--------"

type EcdsaPrivateKey struct {
	Hex string

	*ecdsa.PrivateKey `env:"-"`
}

func (k *EcdsaPrivateKey) Init() error {
	sk, err := crypto.HexToECDSA(k.Hex)
	if err != nil {
		return err
	}
	k.PrivateKey = sk
	return nil
}

func (k *EcdsaPrivateKey) SecurityString() string {
	return MaskedPrivateKey
}

type EcdsaPublicKey struct {
	Hex string

	*ecdsa.PublicKey `env:"-"`
}

func (k *EcdsaPublicKey) Init() error {
	raw, err := hex.DecodeString(k.Hex)
	if err != nil {
		return err
	}

	pk, err := crypto.UnmarshalPubkey(raw)
	if err != nil {
		return err
	}
	k.PublicKey = pk
	return nil
}
