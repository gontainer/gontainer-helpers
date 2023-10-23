package container_test

import (
	"context"
	"strings"
	"testing"

	"github.com/gontainer/gontainer-helpers/v2/container"
	"github.com/stretchr/testify/assert"
)

type myWrappedContainer struct {
	*container.Container
}

func newMyWrappedContainer() *myWrappedContainer {
	return &myWrappedContainer{
		Container: container.New(),
	}
}

func TestContextWithContainer(t *testing.T) {
	// Make sure that approach (embedding struct implementing an interface with unexported methods)
	// works in all GO's versions

	t.Run("Wrapped *Container", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		container.ContextWithContainer(ctx, newMyWrappedContainer())
	})
	t.Run("Invalid context", func(t *testing.T) {
		defer func() {
			assert.True(t, strings.Contains(recover().(string), "`ctx = container.ContextWithContainer(ctx, c)`"))
		}()
		c := container.New()
		_, _ = c.GetInContext(context.Background(), "service")
	})
}
