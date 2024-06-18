package address

import "testing"

func TestAddress(t *testing.T) {
	addr := HexToAddress("0x00000000000000000000000031c3785bebe03cc5ba691c486d6d1cdf8bb438c4")
	if addr.String() != "0x31c3785bEBe03cc5bA691c486d6D1CDF8BB438c4" {
		t.Fatalf("address error got %s, want 0x31c3785bebe03cc5ba691c486d6d1cdf8bb438c4", addr.String())
	}
}
