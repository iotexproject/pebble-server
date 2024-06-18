package abi

import (
	"math/big"
	"testing"

	"github.com/machinefi/sprout-pebble-sequencer/cmd/wasm/tinygoeth/common"
)

func TestEventUnpack(t *testing.T) {
	event := `FirmwareUpdated(string name, string version, string uri, string avatar)`
	e, err := NewEvent(event)
	if err != nil {
		t.Fatal(err)
	}
	data := `0x000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000160000000000000000000000000000000000000000000000000000000000000000c47726176656c5f315f305f3200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005312e302e32000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003e68747470733a2f2f706562626c652d6f74612e73332e61702d656173742d312e616d617a6f6e6177732e636f6d2f47726176656c5f315f305f322e62696e0000000000000000000000000000000000000000000000000000000000000000004d68747470733a2f2f73746f726167656170692e666c65656b2e636f2f7261756c6c656e636861692d7465616d2d6275636b65742f313632323638353030323630392d617661746172312e706e6700000000000000000000000000000000000000`
	v, err := e.Unpack(common.FromHex(data))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("ID: %s", e.ID)
	t.Logf("%+v", v)
}

func TestEventPaidUnpack(t *testing.T) {
	event := `Paid(address indexed from, address indexed to, uint256 amount, uint256 timestamp, uint256 balance)`
	e, err := NewEvent(event)
	if err != nil {
		t.Fatal(err)
	}
	data := `000000000000000000000000000000000000000000000000023999776f8c8000000000000000000000000000000000000000000000000000000000006406a93e0000000000000000000000000000000000000000000000004c1bfb0bf35ed000`
	v, err := e.Unpack(common.FromHex(data))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("ID: %s", e.ID)
	amount := v[0].(*big.Int)
	t.Logf("amount: %s", amount.String())
	timestamp := v[1].(*big.Int)
	t.Logf("timestamp: %s", timestamp.String())
	balance := v[2].(*big.Int)
	t.Logf("balance: %s", balance.String())
	t.Logf("%+v", v)
}

func TestEventSig(t *testing.T) {
	tests := []struct {
		event string
		want  string
	}{
		{
			event: `FirmwareUpdated(string name, string version, string uri, string avatar)`,
			want:  "0x3e7cc89e0f3e642577aa6cf551ebdb03ac285acb710d0233a96bd2319b7e759f",
		},
		{
			event: `FirmwareRemoved(string name)`,
			want:  "0xac33359a95c4778630007ee3bba020f5941f816296c819feb1c95bc90de05a1b",
		},
		{
			event: `Proposal(string imei, address owner, address device, string name, string avatar)`,
			want:  "0x9ffdf0136249d99680088653555755221714868b4f7ca1ff7d8523e3bef1dc4a",
		},
		{
			event: `Confirm(string imei, address owner, address device, uint32 channel)`,
			want:  "0xd5b2cf831feccbfa01ac1987c17d405937ce2aedcb9a9c9efc88853b2a1e2c32",
		},
		{
			event: `Firmware(string imei, string app)`,
			want:  "0x79a23f77451563d737c2d12dba4995c66f9c63fe6703e1b843d634ed3529c12d",
		},
		{
			event: `Config(string imei, string config)`,
			want:  "0x122b97366a660eeb51ccc40eece673685afe30b789c7d940802486bb26a293c6",
		},
		{
			event: `Remove(string imei, address owner)`,
			want:  "0x33f0e5ba8079ed962d5166bdb2180f83068317c73732f76dc437da45bb69ac11",
		},
		{
			event: `Deposit(address indexed to, uint256 amount, uint256 balance)`,
			want:  "0x90890809c654f11d6e72a28fa60149770a0d11ec6c92319d6ceb2bb0a4ea1a15",
		},
		{
			event: `Withdraw(address indexed from, address indexed to, uint256 amount, uint256 balance)`,
			want:  "0xf341246adaac6f497bc2a656f546ab9e182111d630394f0c57c710a59a2cb567",
		},
		{
			event: `Paid(address indexed from, address indexed to, uint256 amount, uint256 timestamp, uint256 balance)`,
			want:  "0xb9fb64ccf647f3e7ba45742b97b6b8e464a822c67817276accb7b1f905d292a2",
		},
		{
			event: `Updated(address owner, string name, string avatar)`,
			want:  "0xeee8917d51964969088b5c14a664dc2d21084932a50ce74fa8ac013403bc4212",
		},
	}
	for _, tt := range tests {
		e, err := NewEvent(tt.event)
		if err != nil {
			t.Fatal(err)
		}
		if got := e.ID.String(); got != tt.want {
			t.Fatalf("evnet %s, got %s, want %s", tt.event, got, tt.want)
		}
	}
}
