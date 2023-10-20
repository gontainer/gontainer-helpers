package container_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gontainer/gontainer-helpers/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_container_concurrency(t *testing.T) {
	const max = 100

	t.Run("ContextualBag", func(t *testing.T) {
		// Values A, B, C, and D must be equal in the same context but they must differ in different contexts.

		type Numbers struct {
			A, B, C, D int64
		}

		c := container.New()

		next := int64(0)
		svcNextInt := container.NewService()
		svcNextInt.SetConstructor(func() int64 {
			return atomic.AddInt64(&next, 1)
		})
		svcNextInt.ScopeContextual()
		c.OverrideService("nextInt", svcNextInt)

		svcNum := container.NewService()
		svcNum.
			SetConstructor(func() *Numbers { return &Numbers{} }).
			SetField("A", container.NewDependencyService("nextInt")).
			SetField("B", container.NewDependencyService("nextInt")).
			SetField("C", container.NewDependencyService("nextInt")).
			SetField("D", container.NewDependencyService("nextInt"))

		for i := 0; i < max; i++ {
			c.OverrideService(fmt.Sprintf("numbers%d", i), svcNum)
		}

		wg := sync.WaitGroup{}
		for i := 0; i < max; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				tmp, err := c.Get(fmt.Sprintf("numbers%d", i))
				assert.NoError(t, err)

				nums := tmp.(*Numbers)
				assert.Equal(t, nums.A, nums.B)
				assert.Equal(t, nums.A, nums.C)
				assert.Equal(t, nums.A, nums.D)

			}(i)
		}
		wg.Wait()

		tmp, err := c.Get("numbers0")
		require.NoError(t, err)
		nums := tmp.(*Numbers)
		assert.Equal(t, nums.A, int64(max+1))
		assert.Equal(t, nums.B, int64(max+1))
		assert.Equal(t, nums.C, int64(max+1))
		assert.Equal(t, nums.D, int64(max+1))
	})

	t.Run("Cache for params", func(t *testing.T) {
		// fatal error: concurrent map writes

		c := container.New()

		for i := 0; i < max; i++ {
			c.OverrideParam(fmt.Sprintf("param%d", i), container.NewDependencyValue(123))
		}

		wg := sync.WaitGroup{}
		wg.Add(max)
		for j := 0; j < max; j++ {
			i := j
			go func() {
				defer wg.Done()
				_, _ = c.GetParam(fmt.Sprintf("param%d", i))
			}()
		}
		wg.Wait()
	})

	t.Run("Cache for shared services", func(t *testing.T) {
		// make sure we don't have the following errors
		// when we cache shared services
		// fatal error: concurrent map read and map write
		// fatal error: concurrent map writes

		c := container.New()

		for i := 0; i < max; i++ {
			s := container.NewService()
			s.SetValue(struct{}{})
			s.ScopeShared()
			c.OverrideService(fmt.Sprintf("service%d", i), s)
		}

		wg := sync.WaitGroup{}
		wg.Add(max)
		for j := 0; j < max; j++ {
			i := j
			go func() {
				defer wg.Done()
				_, _ = c.Get(fmt.Sprintf("service%d", i))
			}()
		}
		wg.Wait()
	})

	t.Run("Cache for shared services with context", func(t *testing.T) {
		// make sure we don't have the following errors
		// when we cache contextual services in context
		// fatal error: concurrent map read and map write
		// fatal error: concurrent map writes

		c := container.New()

		for i := 0; i < max; i++ {
			s := container.NewService()
			s.SetValue(struct{}{})
			s.ScopeContextual()
			c.OverrideService(fmt.Sprintf("service-context%d", i), s)
		}

		ctx := container.ContextWithContainer(context.Background(), c)

		wg := sync.WaitGroup{}
		wg.Add(max)
		for j := 0; j < max; j++ {
			i := j
			go func() {
				defer wg.Done()
				_, _ = c.GetInContext(ctx, fmt.Sprintf("service-context%d", i))
			}()
		}
		wg.Wait()
	})

	t.Run("OverrideService", func(t *testing.T) {
		// fatal error: concurrent map writes

		c := container.New()

		wg := sync.WaitGroup{}
		wg.Add(max)
		for j := 0; j < max; j++ {
			i := j
			go func() {
				defer wg.Done()

				s := container.NewService()
				s.SetValue(struct{}{})
				c.OverrideService(fmt.Sprintf("service%d", i), s)
			}()
		}
		wg.Wait()
	})

	t.Run("AddDecorator", func(t *testing.T) {
		// race detected during execution of test

		c := container.New()

		wg := sync.WaitGroup{}
		wg.Add(max)
		for j := 0; j < max; j++ {
			i := j
			go func() {
				defer wg.Done()
				c.AddDecorator(
					fmt.Sprintf("tag%d", i),
					func(p container.DecoratorPayload) (any, error) {
						// it does nothing
						return p.Service, nil
					},
				)
			}()
		}
		wg.Wait()
	})

	t.Run("All", func(t *testing.T) {
		c := container.New()
		name := container.NewService()
		name.SetValue("Johnny")
		c.OverrideService("name", name)

		newService := func(tag string) container.Service {
			s := container.NewService()
			s.SetConstructor(func() any {
				return struct {
					Name string
				}{}
			})
			s.SetField("Name", container.NewDependencyService("name"))
			s.Tag(tag, 0)
			return s
		}

		for i := 0; i < max; i++ {
			c.OverrideService(fmt.Sprintf("service%d", i), newService("tag"))

			sCtx := newService("tag-context")
			sCtx.ScopeContextual()
			c.OverrideService(fmt.Sprintf("service-context%d", i), sCtx)
		}

		ctx := container.ContextWithContainer(context.Background(), c)

		wg := sync.WaitGroup{}
		wg.Add(max * 9)
		for i := 0; i < max; i++ {
			n := fmt.Sprintf("service%d", i)
			nCtx := fmt.Sprintf("service-context%d", i)

			go func() {
				defer wg.Done()

				c.OverrideService(n, newService("tag"))
			}()

			go func() {
				defer wg.Done()

				s := newService("tag-context")
				s.ScopeContextual()
				c.OverrideService(nCtx, s)
			}()

			go func() {
				defer wg.Done()

				_, err := c.Get(n)
				assert.NoError(t, err)
			}()

			go func() {
				defer wg.Done()

				assert.NoError(t, c.CircularDeps())
			}()

			go func() {
				defer wg.Done()

				c.AddDecorator("tag", func(p container.DecoratorPayload) any {
					return p.Service
				})
			}()

			go func() {
				defer wg.Done()

				assert.True(t, c.IsTaggedBy(n, "tag"))
			}()

			go func() {
				defer wg.Done()

				tagged, err := c.GetTaggedBy("tag")

				assert.NoError(t, err)
				assert.Len(t, tagged, max)
			}()

			go func() {
				defer wg.Done()

				_, err := c.GetInContext(ctx, nCtx)
				assert.NoError(t, err)
			}()

			go func() {
				defer wg.Done()

				tagged, err := c.GetTaggedByInContext(ctx, "tag-context")

				assert.NoError(t, err)
				assert.Len(t, tagged, max)
			}()
		}
		wg.Wait()
	})
}
