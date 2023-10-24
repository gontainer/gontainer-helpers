package container_test

import (
	"context"
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
	t.Run("Wrapped *Container", func(t *testing.T) {
		// Make sure that approach (embedding struct implementing an interface with unexported methods)
		// works in all GO's versions
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		container.ContextWithContainer(ctx, newMyWrappedContainer())
	})
	t.Run("Invalid input", func(t *testing.T) {
		t.Run("ctx.Done() == nil", func(t *testing.T) {
			defer func() {
				assert.Equal(t, `ctx.Done() == nil: a receive from a nil channel blocks forever`, recover())
			}()
			container.ContextWithContainer(context.Background(), container.New())
		})
		t.Run("container == nil", func(t *testing.T) {
			defer func() {
				assert.Equal(t, `nil container`, recover())
			}()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			container.ContextWithContainer(ctx, nil)
		})
		t.Run("context == nil", func(t *testing.T) {
			defer func() {
				assert.Equal(t, `nil context`, recover())
			}()
			container.ContextWithContainer(nil, nil)
		})
	})
}
