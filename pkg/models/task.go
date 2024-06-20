package models

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	ProjectID      uint64 `gorm:"index:task_fetch,not null"`
	InternalTaskID string `gorm:"index:internal_task_id,not null"`
	MessageIDs     []byte `gorm:"not null"`
	Signature      string `gorm:"not null,default:''"`
}

func (t *Task) Sign(sk *ecdsa.PrivateKey, msg *Message) error {
	if msg.ProjectID != t.ProjectID {
		return errors.New("unmatched project id")
	}

	buf := bytes.NewBuffer(nil)
	_ = binary.Write(buf, binary.BigEndian, uint64(t.ID))
	_ = binary.Write(buf, binary.BigEndian, t.ProjectID)
	_, _ = buf.WriteString(msg.ClientID)
	_, _ = buf.Write(crypto.Keccak256Hash(msg.Data).Bytes())

	h := crypto.Keccak256Hash(buf.Bytes())
	sig, err := crypto.Sign(h.Bytes(), sk)
	if err != nil {
		return err
	}
	t.Signature = hexutil.Encode(sig)
	return nil
}
