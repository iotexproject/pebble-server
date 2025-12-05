// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package pebble

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// PebbleMetaData contains all meta data concerning the Pebble contract.
var PebbleMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"string\",\"name\":\"imei\",\"type\":\"string\"},{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"},{\"internalType\":\"uint32\",\"name\":\"timestamp\",\"type\":\"uint32\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"gas\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"channel\",\"type\":\"uint32\"}],\"name\":\"confirm\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// PebbleABI is the input ABI used to generate the binding from.
// Deprecated: Use PebbleMetaData.ABI instead.
var PebbleABI = PebbleMetaData.ABI

// Pebble is an auto generated Go binding around an Ethereum contract.
type Pebble struct {
	PebbleCaller     // Read-only binding to the contract
	PebbleTransactor // Write-only binding to the contract
	PebbleFilterer   // Log filterer for contract events
}

// PebbleCaller is an auto generated read-only Go binding around an Ethereum contract.
type PebbleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PebbleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PebbleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PebbleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PebbleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PebbleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PebbleSession struct {
	Contract     *Pebble           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PebbleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PebbleCallerSession struct {
	Contract *PebbleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// PebbleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PebbleTransactorSession struct {
	Contract     *PebbleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PebbleRaw is an auto generated low-level Go binding around an Ethereum contract.
type PebbleRaw struct {
	Contract *Pebble // Generic contract binding to access the raw methods on
}

// PebbleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PebbleCallerRaw struct {
	Contract *PebbleCaller // Generic read-only contract binding to access the raw methods on
}

// PebbleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PebbleTransactorRaw struct {
	Contract *PebbleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPebble creates a new instance of Pebble, bound to a specific deployed contract.
func NewPebble(address common.Address, backend bind.ContractBackend) (*Pebble, error) {
	contract, err := bindPebble(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Pebble{PebbleCaller: PebbleCaller{contract: contract}, PebbleTransactor: PebbleTransactor{contract: contract}, PebbleFilterer: PebbleFilterer{contract: contract}}, nil
}

// NewPebbleCaller creates a new read-only instance of Pebble, bound to a specific deployed contract.
func NewPebbleCaller(address common.Address, caller bind.ContractCaller) (*PebbleCaller, error) {
	contract, err := bindPebble(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PebbleCaller{contract: contract}, nil
}

// NewPebbleTransactor creates a new write-only instance of Pebble, bound to a specific deployed contract.
func NewPebbleTransactor(address common.Address, transactor bind.ContractTransactor) (*PebbleTransactor, error) {
	contract, err := bindPebble(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PebbleTransactor{contract: contract}, nil
}

// NewPebbleFilterer creates a new log filterer instance of Pebble, bound to a specific deployed contract.
func NewPebbleFilterer(address common.Address, filterer bind.ContractFilterer) (*PebbleFilterer, error) {
	contract, err := bindPebble(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PebbleFilterer{contract: contract}, nil
}

// bindPebble binds a generic wrapper to an already deployed contract.
func bindPebble(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PebbleMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Pebble *PebbleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Pebble.Contract.PebbleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Pebble *PebbleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Pebble.Contract.PebbleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Pebble *PebbleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Pebble.Contract.PebbleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Pebble *PebbleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Pebble.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Pebble *PebbleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Pebble.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Pebble *PebbleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Pebble.Contract.contract.Transact(opts, method, params...)
}

// Confirm is a paid mutator transaction binding the contract method 0x2409556e.
//
// Solidity: function confirm(string imei, address _owner, uint32 timestamp, bytes signature, uint256 gas, uint32 channel) returns()
func (_Pebble *PebbleTransactor) Confirm(opts *bind.TransactOpts, imei string, _owner common.Address, timestamp uint32, signature []byte, gas *big.Int, channel uint32) (*types.Transaction, error) {
	return _Pebble.contract.Transact(opts, "confirm", imei, _owner, timestamp, signature, gas, channel)
}

// Confirm is a paid mutator transaction binding the contract method 0x2409556e.
//
// Solidity: function confirm(string imei, address _owner, uint32 timestamp, bytes signature, uint256 gas, uint32 channel) returns()
func (_Pebble *PebbleSession) Confirm(imei string, _owner common.Address, timestamp uint32, signature []byte, gas *big.Int, channel uint32) (*types.Transaction, error) {
	return _Pebble.Contract.Confirm(&_Pebble.TransactOpts, imei, _owner, timestamp, signature, gas, channel)
}

// Confirm is a paid mutator transaction binding the contract method 0x2409556e.
//
// Solidity: function confirm(string imei, address _owner, uint32 timestamp, bytes signature, uint256 gas, uint32 channel) returns()
func (_Pebble *PebbleTransactorSession) Confirm(imei string, _owner common.Address, timestamp uint32, signature []byte, gas *big.Int, channel uint32) (*types.Transaction, error) {
	return _Pebble.Contract.Confirm(&_Pebble.TransactOpts, imei, _owner, timestamp, signature, gas, channel)
}
