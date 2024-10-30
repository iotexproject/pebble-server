package contexts_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotexproject/pebble-server/contexts"
)

func TestWhiteList_NeedHandle(t *testing.T) {
	r := require.New(t)
	wl := contexts.WhiteList{"112", "212"}

	r.True(wl.NeedHandle("112"))
	r.True(wl.NeedHandle("212"))
	r.False(wl.NeedHandle("119"))
	r.True((contexts.WhiteList{}).NeedHandle("any"))
	r.True(wl.NeedHandle("any___0"))
	r.True(wl.NeedHandle("any___1"))
	r.True(wl.NeedHandle("any___2"))
	r.True(wl.NeedHandle("any___3"))
	r.True(wl.NeedHandle("any___4"))
	r.False(wl.NeedHandle("any___5"))
}
