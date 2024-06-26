package main

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/machinefi/w3bstream-wasm-golang-sdk/log"
	"github.com/machinefi/w3bstream-wasm-golang-sdk/stream"
	"github.com/tidwall/gjson"

	"github.com/machinefi/sprout-pebble-sequencer/cmd/wasm/tinygoeth/abi"
	"github.com/machinefi/sprout-pebble-sequencer/cmd/wasm/tinygoeth/address"
)

// main is required for TinyGo to compile to Wasm.
func main() {}

/*
type message struct {
	IMEI        string `json:"imei"`
	Owner       string `json:"owner"`
	Timestamp   uint32 `json:"timestamp"`
	Signature   string `json:"signature"`
	GasLimit    string `json:"gasLimit"`
	DataChannel uint32 `json:"dataChannel"`
}
*/

//export start
func onStart(rid uint32) int32 {
	res, err := stream.GetDataByRID(rid)
	if err != nil {
		log.Log("error: " + err.Error())
		return -1
	}
	log.Log(fmt.Sprintf("get resource %v: `%s`", rid, string(res)))

	log.Log("wasm get datas from json: " + gjson.Get(string(res), "datas").String())
	datas := gjson.Get(string(res), "datas").Array()
	imei := gjson.Get(datas[0].String(), "imei").String()
	owner := address.HexToAddress(gjson.Get(datas[0].String(), "owner").String())
	timestamp := uint32(gjson.Get(datas[0].String(), "timestamp").Uint())
	signature, _ := hex.DecodeString(gjson.Get(datas[0].String(), "signature").String())
	gasLimit, _ := new(big.Int).SetString(gjson.Get(datas[0].String(), "gasLimit").String(), 10)
	dataChannel := uint32(gjson.Get(datas[0].String(), "dataChannel").Uint())

	/*https://github.com/iotexproject/pebble-contracts/blob/1a1c91a287317d8c068edb571149aedb10c0b754/contracts/PebbleImpl.sol
		confirm(
	        string memory imei,
	        address _owner,
	        uint32 timestamp,
	        bytes memory signature,
	        uint256 gas,
	        uint32 channel
	    )
	*/
	method, err := abi.NewMethod("confirm(string,address,uint32,bytes,uint256,uint32)")
	if err != nil {
		log.Log(fmt.Sprintf("abi.NewMethod error: %s", err.Error()))
		return -1
	}

	data, err := method.Pack(imei, owner, timestamp, signature, gasLimit, dataChannel)
	if err != nil {
		log.Log(fmt.Sprintf("pack error: %s", err.Error()))
		return -1
	}
	log.Log("rawData: " + hex.EncodeToString(data))
	stream.SetBytesByRID(rid, data)
	return int32(rid)
}
