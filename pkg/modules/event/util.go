package event

import (
	"bytes"
	"context"
	"crypto/elliptic"
	"encoding/binary"
	"hash"
	"io"
	"sort"
	"strings"

	"github.com/dustinxie/ecc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
	"github.com/xoctopus/x/misc/must"
	"golang.org/x/crypto/sha3"
	"gorm.io/gorm/schema"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/contexts"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
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

type TxHash struct {
	hash common.Hash
}

func (h *TxHash) Set(log *types.Log) {
	h.hash = log.TxHash
}

func (h TxHash) Hash() common.Hash {
	return h.hash
}

func (h TxHash) String() string {
	return h.hash.String()
}

func FetchDeviceByIMEI(ctx context.Context, imei string) (*models.Device, error) {
	return nil, nil
}

func FetchFirmwareByID(ctx context.Context, id string) (*models.App, error) {
	return nil, nil
}

type Assigner struct {
	Name string
	V    any
}

func Update(ctx context.Context, t schema.Tabler, kvs ...*Assigner) error {
	db := must.BeTrueV(contexts.DatabaseFromContext(ctx))
	q, vs := BuildUpdateQuery(t.TableName(), kvs...)
	if q == nil {
		return nil
	}
	return db.Exec(*q, vs...).Error
}

func UpsertOnConflictUpdateOthers(ctx context.Context, t schema.Tabler, cs []string, kvs ...*Assigner) error {
	db := must.BeTrueV(contexts.DatabaseFromContext(ctx))
	q, vs := BuildUpsertOnConflictUpdateOthersQuery(t.TableName(), cs, kvs...)
	if q == nil {
		return nil
	}
	return db.Exec(*q, vs...).Error
}

func UpsertOnConflictDoNothing(ctx context.Context, t schema.Tabler, cs []string, kvs ...*Assigner) error {
	return nil
}

func BuildUpdateQuery(table string, kvs ...*Assigner) (*string, []any) {
	if len(kvs) == 0 {
		return nil, nil
	}

	vs := make([]any, 0, len(kvs))

	assignments := make([]string, 0, len(kvs))
	for _, pair := range kvs {
		if pair.Name == "" || pair.V == nil {
			continue
		}
		assignments = append(assignments, pair.Name+"=?")
		vs = append(vs, pair.V)
	}

	if len(vs) == 0 {
		return nil, nil
	}

	q := `UPDATE ` + table + ` SET ` + strings.Join(assignments, ",")

	return &q, vs
}

func BuildUpsertOnConflictUpdateOthersQuery(table string, conflicts []string, kvs ...*Assigner) (*string, []any) {
	if len(kvs) == 0 {
		return nil, nil
	}

	vs := make([]any, 0, len(kvs)+len(conflicts))
	vm := map[string]any{}

	valuenames := make([]string, 0, len(kvs))
	valueholders := make([]string, 0, len(kvs))
	for _, pair := range kvs {
		if pair.Name == "" || pair.V == nil {
			continue
		}
		valuenames = append(valuenames, pair.Name)
		if _, ok := vm[pair.Name]; ok {
			panic(errors.Errorf("assigner name %s duplicated", pair.Name))
		}
		vm[pair.Name] = pair.V
		vs = append(vs, pair.V)
		valueholders = append(valueholders, "?")
	}

	if len(valuenames) == 0 {
		return nil, nil
	}

	conflictnames := make([]string, 0, len(conflicts))
	conflictkv := map[string]struct{}{}
	for _, name := range conflicts {
		_, ok := vm[name]
		if !ok {
			panic(errors.Errorf("conflict name %s not in upsert list", name))
		}
		conflictnames = append(conflictnames, name)
		conflictkv[name] = struct{}{}
	}

	otherassignments := make([]string, 0)
	for name, v := range vm {
		if _, ok := conflictkv[name]; !ok && name != "created_at" {
			otherassignments = append(otherassignments, name+"=?")
			vs = append(vs, v)
		}
	}
	sort.Slice(otherassignments, func(i, j int) bool {
		return otherassignments[i] < otherassignments[j]
	})

	q := `INSERT INTO ` + table + ` (` + strings.Join(valuenames, ",") + `) VALUES (` +
		strings.Join(valueholders, ",") + `)`
	if len(conflicts) > 0 {
		q += ` ON CONFLICT (` + strings.Join(conflictnames, ",") + `) DO UPDATE SET ` +
			strings.Join(otherassignments, ",")
	}
	return &q, vs
}
