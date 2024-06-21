package models_test

import (
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
	"gorm.io/gorm"
	"testing"
)

func TestTask_Sign(t *testing.T) {
	task := &models.Task{
		Model:          gorm.Model{ID: 10},
		ProjectID:      100,
		InternalTaskID: uuid.NewString(),
		MessageIDs:     []byte(""),
		Signature:      "",
	}
	sk, err := crypto.HexToECDSA("dbfe03b0406549232b8dccc04be8224fcc0afa300a33d4f335dcfdfead861c85")
	if err != nil {
		t.Fatal(err)
	}
	err = task.Sign(sk, &models.Message{ProjectID: 100, Data: []byte("any")})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(task.Signature)
}
