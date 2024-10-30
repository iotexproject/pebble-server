package event

import (
	"bytes"
	"crypto/elliptic"
	"hash"
	"io"

	"github.com/dustinxie/ecc"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

type CanValidateSignature interface {
	Address() common.Address
	Hash() []byte
	Signature() []byte
	Validate() bool
}

type SignatureValidator struct {
	addr common.Address
	hash []byte
	sig  []byte
}

func (sv *SignatureValidator) Address() common.Address { return sv.addr }

func (sv *SignatureValidator) Hash() []byte { return sv.hash }

func (sv *SignatureValidator) Signature() []byte { return sv.sig }

func (sv *SignatureValidator) Validate() bool {
	for i := 0; i < 4; i++ {
		sv.sig[64] = byte(i)
		pk, err := ecc.RecoverPubkey("P-256k1", sv.hash, sv.sig)
		if err != nil {
			continue
		}
		if pk != nil && pk.X != nil && pk.Y != nil &&
			ecc.P256k1().IsOnCurve(pk.X, pk.Y) {
			raw := elliptic.Marshal(ecc.P256k1(), pk.X, pk.Y)
			raw = raw[1:]

			// Keccak256
			b := make([]byte, 32)
			s := sha3.NewLegacyKeccak256()
			s.(hash.Hash).Write(raw)
			_, _ = s.(io.Reader).Read(b)

			if bytes.Equal(common.BytesToAddress(b[12:]).Bytes(), sv.addr.Bytes()) {
				return true
			}
		}
	}
	return false
}
