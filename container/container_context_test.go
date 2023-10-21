package container_test

import (
	"context"
	"strings"
	"testing"

	"github.com/gontainer/gontainer-helpers/container"
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

type myContainerWithOverriddenFunc struct {
	*container.Container
}

func newMyContainerWithOverriddenFunc() *myContainerWithOverriddenFunc {
	return &myContainerWithOverriddenFunc{
		Container: container.New(),
	}
}

func (*myContainerWithOverriddenFunc) getContainerID() int { //nolint:all
	panic("it should not be invoked")
}

func TestContextWithContainer(t *testing.T) {
	// Make sure that approach (embedding struct implementing an interface with unexported methods)
	// works in all GO's versions

	t.Run("SuperContainer", func(t *testing.T) {
		container.ContextWithContainer(context.Background(), container.New())
	})
	t.Run("Wrapped Container", func(t *testing.T) {
		container.ContextWithContainer(context.Background(), newMyWrappedContainer())
	})
	t.Run("Wrapped Container with overridden func", func(t *testing.T) {
		container.ContextWithContainer(context.Background(), newMyContainerWithOverriddenFunc())
	})
	t.Run("Invalid context", func(t *testing.T) {
		defer func() {
			assert.True(t, strings.Contains(recover().(string), "`ctx = container.ContextWithContainer(ctx, c)`"))
		}()
		c := container.New()
		_, _ = c.GetInContext(context.Background(), "service")
	})
}
