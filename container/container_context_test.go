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

func (*myContainerWithOverriddenFunc) getContainerID() int {
	panic("it should not be invoked")
}

func TestContextWithContainer(t *testing.T) {
	// Make sure the interface with unexported func works on all GO versions,
	// it won't cause errors when there is a conflict name
	// in a struct that embeds the type that implements the interface.

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
