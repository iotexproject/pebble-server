package models_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/iotexproject/pebble-server/models"
)

func TestTask_Sign(t *testing.T) {
	task := &models.Task{
		Model:          gorm.Model{ID: 10},
		ProjectID:      100,
		InternalTaskID: uuid.NewString(),
		MessageIDs:     []byte(""),
		Signature:      "",
	}

	// sign
	sk, err := crypto.HexToECDSA("dbfe03b0406549232b8dccc04be8224fcc0afa300a33d4f335dcfdfead861c85")
	if err != nil {
		t.Fatal(err)
	}
	err = task.Sign(sk, &models.Message{ClientID: "any", ProjectID: 100, Data: []byte("any")})
	if err != nil {
		t.Fatal(err)
	}
}
