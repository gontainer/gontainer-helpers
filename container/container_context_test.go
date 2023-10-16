package container_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
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

func TestContextWithContainer1(t *testing.T) {
	c := container.NewContainer()
	s := container.NewService()
	s.SetValue(5)
	c.OverrideService("five", s)

	fives := new(uint64)
	sixes := new(uint64)
	wg := &sync.WaitGroup{}

	run := func() {
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer func() {
					go wg.Done()
				}()

				v, _ := c.Get("five")
				if v == 5 {
					atomic.AddUint64(fives, 1)
				} else {
					atomic.AddUint64(sixes, 1)
				}
			}()
		}
	}

	run()

	go func() {
		s.SetValue(6)
		c.OverrideService("five", s)
	}()

	run()

	wg.Wait()
	fmt.Println(*fives, *sixes)
}
