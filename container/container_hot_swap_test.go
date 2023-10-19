package container_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gontainer/gontainer-helpers/container"
	"github.com/stretchr/testify/assert"
)

func TestNewContainer_hotSwap(t *testing.T) {
	t.Run("Wait for <-ctx.Done()", func(t *testing.T) {
		c := container.NewContainer()
		s := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		container.ContextWithContainer(ctx, c)
		triggered := false
		c.HotSwap(func(container.MutableContainer) {
			triggered = true
		})
		assert.True(t, triggered)
		assert.GreaterOrEqual(t, time.Since(s), time.Second)
	})
	t.Run("HotSwap", func(t *testing.T) {
		counter := new(uint64)

		s := container.NewService()
		s.SetConstructor(func() Person {
			atomic.AddUint64(counter, 1)
			return Person{Name: "Jane"}
		})

		c := container.NewContainer()
		c.OverrideService("person", s)

		const max = 1000
		wg := sync.WaitGroup{}

		runGoroutines := func() {
			wg.Add(max)
			for i := 0; i < max; i++ {
				go func() {
					defer wg.Done()

					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					ctx = container.ContextWithContainer(ctx, c)
					_, _ = c.GetInContext(ctx, "person")
				}()
			}
		}

		runGoroutines()
		c.HotSwap(func(mc container.MutableContainer) {
			mc.InvalidateServiceCache("person")
		})
		runGoroutines()
		wg.Wait()

		assert.Equal(t, uint64(2), atomic.LoadUint64(counter))
	})
}
