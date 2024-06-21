package models_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/models"
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
	t.Log(task.Signature)

	signature := "0x5889abfda43e0583f22d3e373aa576c64a57471ad14317289784e607baf3699c39a6092b3db6540c0521f2544e675bf8c42dad8855a8c92c5a1bc499d08f31c101"
	// verify
	sig, err := hexutil.Decode(signature)
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.BigEndian, uint64(100))
	binary.Write(buf, binary.BigEndian, 6457)
	buf.WriteString("0xaB4b1C4Ce799DB98fb3A414a82941bF7e07a5E14")
	data := []byte(`{"imei":"103381234567407","owner":"0xF77f8De24194D768012CA1Edd15AeE0B33D919a1","timestamp":1718964297,"signature":"fe0c811da8c4b130a4eeedd411e9ab694c81e5455d87206bba390746889053e323da069e1109b58ab9cfc055349973d5702595b7f136995d3a812bafe06fca9e","gasLimit":"200000","dataChannel":8183}`)
	buf.Write(crypto.Keccak256Hash(data).Bytes())

	h := crypto.Keccak256Hash(buf.Bytes())
	pk, err := crypto.Ecrecover(h.Bytes(), sig)
	if !bytes.Equal(pk, crypto.FromECDSAPub(&sk.PublicKey)) {
		t.Fatal("not equal")
	}
}
