package main

import (
	_ "embed"

	"github.com/ethereum/go-ethereum/common"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/enums"
	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/blockchain"
)

// config reference: https://github.com/iotexproject/pebble-contracts/blob/main/README.md

var (
	//go:embed abis/account_updated.json
	abiAccountUpdated string
	//go:embed abis/firmware_removed.json
	abiFirmwareRemoved string
	//go:embed abis/firmware_updated.json
	abiFirmwareUpdated string
	//go:embed abis/bank_deposit.json
	abiBankDeposit string
	//go:embed abis/bank_paid.json
	abiBankPaid string
	//go:embed abis/bank_withdraw.json
	abiBankWithdraw string
	//go:embed abis/pebble_config.json
	abiPebbleConfig string
	//go:embed abis/pebble_confirm.json
	abiPebbleConfirm string
	//go:embed abis/pebble_firmware.json
	abiPebbleFirmware string
	//go:embed abis/pebble_proposal.json
	abiPebbleProposal string
	//go:embed abis/pebble_remove.json
	abiPebbleRemove string
)

var contracts = []*blockchain.Contract{
	{
		ID:      "IOTX_MAINNET_" + enums.CONTRACT__PEBBLE_ACCOUNT,
		Network: blockchain.NETWORK__IOTX_MAINNET,
		Address: common.HexToAddress("0x189e2ED6EAfBCeAF938d049cf3685828b5493952"),
		Events: []*blockchain.Event{
			{Name: "Updated", ABI: abiAccountUpdated},
		},
	},
	{
		ID:      "IOTX_MAINNET_" + enums.CONTRACT__PEBBLE_FIRMWARE,
		Network: blockchain.NETWORK__IOTX_MAINNET,
		Address: common.HexToAddress("0xC5F406c42C96e68756311Dad49dE99B0f4A1A722"),
		Events: []*blockchain.Event{
			{Name: "FirmwareRemoved", ABI: abiFirmwareRemoved},
			{Name: "FirmwareUpdated", ABI: abiFirmwareUpdated},
		},
	},
	{
		ID:      "IOTX_MAINNET_" + enums.CONTRACT__PEBBLE_BANK,
		Network: blockchain.NETWORK__IOTX_MAINNET,
		Address: common.HexToAddress("0xb86f97D494EEf8c6d618ee2049419eE0Ce843F28"),
		Events: []*blockchain.Event{
			{Name: "Deposit", ABI: abiBankDeposit},
			{Name: "Paid", ABI: abiBankPaid},
			{Name: "Withdraw", ABI: abiBankWithdraw},
		},
	},
	{
		ID:      "IOTX_MAINNET_" + enums.CONTRACT__PEBBLE_DEVICE,
		Network: blockchain.NETWORK__IOTX_MAINNET,
		Address: common.HexToAddress("0xC9D7D9f25b98119DF5b2303ac0Df6b15C982BbF5"),
		Events: []*blockchain.Event{
			{Name: "Config", ABI: abiPebbleConfig},
			{Name: "Confirm", ABI: abiPebbleConfirm},
			{Name: "Firmware", ABI: abiPebbleFirmware},
			{Name: "Proposal", ABI: abiPebbleProposal},
			{Name: "Remove", ABI: abiPebbleRemove},
		},
	},
	{
		ID:      "IOTX_TESTNET_" + enums.CONTRACT__PEBBLE_ACCOUNT,
		Network: blockchain.NETWORK__IOTX_TESTNET,
		Address: common.HexToAddress("0xBC458A041a50BF5abb900C78b7355d54E92FCFBa"),
		Events: []*blockchain.Event{
			{Name: "Updated", ABI: abiAccountUpdated},
		},
	},
	{
		ID:      "IOTX_TESTNET_" + enums.CONTRACT__PEBBLE_FIRMWARE,
		Network: blockchain.NETWORK__IOTX_TESTNET,
		Address: common.HexToAddress("0x93ea43b4548058f4c02925F0725422E6DC81c92d"),
		Events: []*blockchain.Event{
			{Name: "FirmwareRemoved", ABI: abiFirmwareRemoved},
			{Name: "FirmwareUpdated", ABI: abiFirmwareUpdated},
		},
	},
	{
		ID:      "IOTX_TESTNET_" + enums.CONTRACT__PEBBLE_BANK,
		Network: blockchain.NETWORK__IOTX_TESTNET,
		Address: common.HexToAddress("0xd313b3131e238C635f2fE4a84EaDaD71b3ed25fa"),
		Events: []*blockchain.Event{
			{Name: "Deposit", ABI: abiBankDeposit},
			{Name: "Paid", ABI: abiBankPaid},
			{Name: "Withdraw", ABI: abiBankWithdraw},
		},
	},
	{
		ID:      "IOTX_TESTNET_" + enums.CONTRACT__PEBBLE_DEVICE,
		Network: blockchain.NETWORK__IOTX_TESTNET,
		Address: common.HexToAddress("0x1AA325E5144f763a520867c56FC77cC1411430d0"),
		Events: []*blockchain.Event{
			{Name: "Config", ABI: abiPebbleConfig},
			{Name: "Confirm", ABI: abiPebbleConfirm},
			{Name: "Firmware", ABI: abiPebbleFirmware},
			{Name: "Proposal", ABI: abiPebbleProposal},
			{Name: "Remove", ABI: abiPebbleRemove},
		},
	},
}
