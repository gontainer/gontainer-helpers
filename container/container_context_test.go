package container_test

import (
	"context"
	"testing"

	"github.com/gontainer/gontainer-helpers/container"
)

type myWrappedContainer struct {
	*container.SuperContainer
}

func newMyWrappedContainer() *myWrappedContainer {
	return &myWrappedContainer{
		SuperContainer: container.NewSuperContainer(),
	}
}

type myContainerWithOverriddenFunc struct {
	*container.SuperContainer
}

func newMyContainerWithOverriddenFunc() *myContainerWithOverriddenFunc {
	return &myContainerWithOverriddenFunc{
		SuperContainer: container.NewSuperContainer(),
	}
}

func (*myContainerWithOverriddenFunc) getContainerID() int { //nolint:all
	panic("it should not be invoked")
}

func TestContextWithContainer(t *testing.T) {
	// Make sure that approach (embedding struct implementing an interface with unexported methods)
	// works in all GO's versions

	t.Run("SuperContainer", func(t *testing.T) {
		container.ContextWithContainer(context.Background(), container.NewSuperContainer())
	})
	t.Run("Wrapped container", func(t *testing.T) {
		container.ContextWithContainer(context.Background(), newMyWrappedContainer())
	})
	t.Run("Wrapped container with overridden func", func(t *testing.T) {
		container.ContextWithContainer(context.Background(), newMyContainerWithOverriddenFunc())
	})
}
