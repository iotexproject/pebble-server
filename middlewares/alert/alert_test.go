package alert_test

import (
	"testing"

	"github.com/machinefi/sprout-pebble-sequencer/pkg/middlewares/alert"
)

func TestLarkAlert_Push(t *testing.T) {
	n := &alert.LarkAlert{
		URL:     "https://open.larksuite.com/open-apis/bot/v2/hook/f8d7cd45-4b45-40fe-9635-5e2f85e19155",
		Secret:  "vztL7BIOyDw10XEd9H5B6",
		Env:     "dev",
		Project: "unit test",
		Version: "v0.0.1",
	}
	n.Init()

	if err := n.Push("title", "content"); err != nil {
		t.Fatal(err)
	}
}
