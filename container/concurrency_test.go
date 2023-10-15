package container_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/gontainer/gontainer-helpers/container"
	"github.com/stretchr/testify/assert"
)

func Test_paramContainer_concurrency(t *testing.T) {
	const max = 100

	t.Run("Cache for shared params", func(t *testing.T) {
		// fatal error: concurrent map writes

		c := container.NewParamContainer()

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
}

func Test_container_concurrency(t *testing.T) {
	const max = 100

	t.Run("Cache for shared services", func(t *testing.T) {
		// make sure we don't have the following errors
		// when we cache shared services
		// fatal error: concurrent map read and map write
		// fatal error: concurrent map writes

		c := container.NewContainer()

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

		c := container.NewContainer()

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
				_, _ = c.GetWithContext(ctx, fmt.Sprintf("service-context%d", i))
			}()
		}
		wg.Wait()
	})

	t.Run("OverrideService", func(t *testing.T) {
		// fatal error: concurrent map writes

		c := container.NewContainer()

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

		c := container.NewContainer()

		wg := sync.WaitGroup{}
		wg.Add(max)
		for j := 0; j < max; j++ {
			i := j
			go func() {
				defer wg.Done()
				c.AddDecorator(
					fmt.Sprintf("tag%d", i),
					func(p container.DecoratorPayload) (interface{}, error) {
						// it does nothing
						return p.Service, nil
					},
				)
			}()
		}
		wg.Wait()
	})

	t.Run("All", func(t *testing.T) {
		c := container.NewContainer()
		n := container.NewService()
		n.SetValue("Johnny")
		c.OverrideService("name", n)

		newService := func(tag string) container.Service {
			s := container.NewService()
			s.SetConstructor(func() interface{} {
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

			go func() {
				defer wg.Done()

				c.OverrideService(n, newService("tag"))
			}()

			go func() {
				defer wg.Done()

				var s interface{}
				assert.NoError(t, c.CopyServiceTo(n, &s))
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

				c.AddDecorator("tag", func(p container.DecoratorPayload) interface{} {
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
				_, _ = c.GetWithContext(ctx, fmt.Sprintf("service-context%d", i))
			}()

			go func() {
				defer wg.Done()

				tagged, err := c.GetTaggedByWithContext(ctx, "tag-context")

				assert.NoError(t, err)
				assert.Len(t, tagged, max)
			}()
		}
		wg.Wait()
	})
}
