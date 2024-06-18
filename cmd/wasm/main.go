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
	// owner := []byte(datas[0].Get("owner").String())
	owner := []byte(gjson.Get(datas[0].String(), "owner").String())
	// timestamp := uint32(datas[0].Get("timestamp").Uint())
	timestamp := uint32(gjson.Get(datas[0].String(), "timestamp").Uint())
	// signature := []byte(datas[0].Get("signature").String())
	signature := []byte(gjson.Get(datas[0].String(), "signature").String())
	// dataChannel := uint32(datas[0].Get("dataChannel").Uint())
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
	data, err := method.Pack(imei, address.BytesToAddress(owner), timestamp, signature, big.NewInt(200000), dataChannel)
	if err != nil {
		log.Log(fmt.Sprintf("pack error: %s", err.Error()))
		return -1
	}
	log.Log("rawData: " + hex.EncodeToString(data))
	stream.SetDataByRID(rid, hex.EncodeToString(data))
	return int32(rid)
}
