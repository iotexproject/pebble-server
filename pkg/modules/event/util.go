package event

import (
	"bytes"
	"context"
	"crypto/elliptic"
	"encoding/binary"
	"encoding/json"
	"hash"
	"io"

	"github.com/dustinxie/ecc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/xoctopus/x/misc/must"
	"golang.org/x/crypto/sha3"
	"gorm.io/gorm/clause"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
)

var gByteOrder = binary.BigEndian

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

type TopicUnmarshaller struct {
	src    any
	topic  []byte
	prefix string
	suffix string
}

func (m *TopicUnmarshaller) Unmarshal() error {
	parts := bytes.Split(m.topic, []byte("/"))
	if len(parts) != 3 {
		return &UnmarshalTopicError{topic: string(m.topic), event: m.src}
	}
	if !bytes.Equal(parts[0], []byte(m.prefix)) || !bytes.Equal(parts[2], []byte(m.suffix)) {
		return &UnmarshalTopicError{topic: string(m.topic), event: m.src}
	}
	if len(parts[1]) == 0 {
		return &UnmarshalTopicError{topic: string(m.topic), event: m.src}
	}
	if setter, ok := m.src.(WithIMEI); ok {
		setter.SetIMEI(string(parts[1]))
	}
	return nil
}

type WithIMEI interface {
	SetIMEI(string)
	GetIMEI() string
}

type IMEI struct {
	imei string
}

func (i *IMEI) SetIMEI(v string) { i.imei = v }

func (i *IMEI) GetIMEI() string { return i.imei }

type CanSetTxHash interface {
	SetTxHash(l *types.Log)
}

type TxHash struct {
	hash common.Hash
}

func (h *TxHash) SetTxHash(log *types.Log) {
	h.hash = log.TxHash
}

func (h TxHash) Hash() common.Hash {
	return h.hash
}

func (h TxHash) String() string {
	return h.hash.String()
}

func UpsertOnConflict(ctx context.Context, m any, conflict string, updates ...string) (any, error) {
	db := must.BeTrueV(contexts.DatabaseFromContext(ctx))

	cond := clause.OnConflict{
		Columns: []clause.Column{{Name: conflict}},
	}
	if len(updates) == 0 {
		cond.DoNothing = true
	} else {
		cond.DoUpdates = clause.AssignmentColumns(updates)
	}
	if err := db.Clauses(cond).Create(m).Error; err != nil {
		return nil, err
	}
	return m, nil
}

func DeleteByPrimary(ctx context.Context, m any, pk any) error {
	db := must.BeTrueV(contexts.DatabaseFromContext(ctx))
	return db.Delete(m, pk).Error
}

func UpdateByPrimary(ctx context.Context, m any, pk any, fields map[string]any) error {
	db := must.BeTrueV(contexts.DatabaseFromContext(ctx))

	if err := FetchByPrimary(ctx, m, pk); err != nil {
		return err
	}

	return db.Model(m).Updates(fields).Error
}

func FetchByPrimary(ctx context.Context, m any, pk any) error {
	db := must.BeTrueV(contexts.DatabaseFromContext(ctx))
	return db.First(m, pk).Error
}

func PublicMqttMessage(ctx context.Context, id, topic string, v any) error {
	mq := must.BeTrueV(contexts.MqttBrokerFromContext(ctx))

	cli, err := mq.NewClient(id, topic)
	if err != nil {
		return err
	}
	defer mq.Close(cli)

	var data any
	switch data.(type) {
	case string:
		data = v
	case []byte:
		data = v
	default:
		data, err = json.Marshal(v)
		if err != nil {
			return err
		}
	}
	return cli.Publish(data)
}
