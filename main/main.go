package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

type queryReq struct {
	DeviceID  string `json:"deviceID"                   binding:"required"`
	Signature string `json:"signature,omitempty"        binding:"required"`
}

func main() {
	sigStr := "0xf48697dfd90cb6d64daff8ba2e9cb089434745924f52c93af4fd8532acd8db11eda477dd906db3cbd4bd366e4a22cdf0f08bcb1307175bf2475c11d9f8ac6afa"
	sig, err := hexutil.Decode(sigStr)
	if err != nil {
		panic(err)
	}
	req := &queryReq{
		DeviceID: "did:io:0x004521b395a8a54c698bb8f29d4934579d8bda4a",
	}
	reqJson, _ := json.Marshal(req)
	h := sha256.Sum256(reqJson)
	fmt.Println(hexutil.Encode(h[:]))
	recoveryID := uint8(0)

	for recoveryID = 0; recoveryID < 2; recoveryID++ {
		v := byte(recoveryID)
		n := append(sig, v)

		//slog.Info("current signature", "signature", hexutil.Encode(ns))
		if a, err := recover(n, h[:]); err == nil {
			slog.Info("recover owner success", "r_id", recoveryID)
			fmt.Println(a.String())
		} else {
			slog.Error("failed", "error", err)
		}
	}

}

func recover(sig, h []byte) (common.Address, error) {
	sigpk, err := crypto.SigToPub(h, sig)
	if err != nil {
		return common.Address{}, errors.Wrapf(err, "failed to recover public key from signature")
	}
	fmt.Println("public key", hexutil.Encode(crypto.FromECDSAPub(sigpk)))
	return crypto.PubkeyToAddress(*sigpk), nil
}
