package contexts_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xoctopus/x/contextx"

	"github.com/iotexproject/pebble-server/contexts"
)

func TestWithDryRun(t *testing.T) {
	root := context.Background()

	t.Run("FromEmpty", func(t *testing.T) {
		dryrun, ok := contexts.DryRun().From(root)
		assert.Equal(t, dryrun, false)
		assert.Equal(t, ok, false)
	})

	t.Run("MustFromEmpty", func(t *testing.T) {
		dryrun := contexts.DryRun().MustFrom(root)
		assert.False(t, dryrun)
	})

	root = contextx.WithContextCompose(
		contexts.DryRun().Compose(true),
	)(root)

	t.Run("FromContext", func(t *testing.T) {
		dryrun, ok := contexts.DryRun().From(root)
		assert.Equal(t, dryrun, true)
		assert.Equal(t, ok, true)
	})
	t.Run("MustFromContext", func(t *testing.T) {
		dryrun := contexts.DryRun().MustFrom(root)
		assert.Equal(t, dryrun, true)
	})
}
