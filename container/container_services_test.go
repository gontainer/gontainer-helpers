package container_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gontainer/gontainer-helpers/v2/container"
	"github.com/gontainer/gontainer-helpers/v2/copier"
	assertErr "github.com/gontainer/gontainer-helpers/v2/grouperror/assert"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainer_GetInContext(t *testing.T) {
	t.Run("Context is done", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		c := container.New()
		ctx = container.ContextWithContainer(ctx, c)

		_, err := c.GetInContext(ctx, "service")
		assert.EqualError(t, err, `GetInContext("service"): ctx.Done() closed: context canceled`)
	})
}

func TestContainer_GetTaggedByInContext(t *testing.T) {
	t.Run("Context is done", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		c := container.New()
		ctx = container.ContextWithContainer(ctx, c)

		_, err := c.GetTaggedByInContext(ctx, "tag")
		assert.EqualError(t, err, `GetTaggedByInContext("tag"): ctx.Done() closed: context canceled`)
	})
}

func TestContainer_executeServiceCalls(t *testing.T) {
	t.Run("Errors", func(t *testing.T) {
		s := container.NewService()
		s.SetValue(struct{}{})
		s.AppendCall("SetName", container.NewDependencyProvider(func() (any, error) {
			return nil, errors.New("could not fetch the name from the config")
		}))
		s.AppendCall("SetAge", container.NewDependencyValue(21))
		s.AppendCall("SetColor", container.NewDependencyValue("red"))
		s.AppendWither("WithLogger", container.NewDependencyValue(log.New(os.Stdout, "", 0)))
		// this call will be ignored, because it's after the error returned by a wither
		s.AppendCall("SetLanguage", container.NewDependencyValue("en"))

		c := container.New()
		c.OverrideService("service", s)

		expected := []string{
			`get("service"): resolve args "SetName": arg #0: cannot call provider func() (interface {}, error): could not fetch the name from the config`,
			`get("service"): call "SetAge": cannot call method (*interface {})."SetAge": invalid func (*interface {})."SetAge"`,
			`get("service"): call "SetColor": cannot call method (*interface {})."SetColor": invalid func (*interface {})."SetColor"`,
			`get("service"): wither "WithLogger": cannot call wither (struct {})."WithLogger": invalid func (struct {})."WithLogger"`,
		}

		svc, err := c.Get("service")
		assert.Nil(t, svc)
		assertErr.EqualErrorGroup(t, err, expected)
	})
}

func TestContainer_createNewService(t *testing.T) {
	t.Run("Error in provider", func(t *testing.T) {
		s := container.NewService()
		s.SetConstructor(func() (any, error) {
			return nil, errors.New("could not create")
		})

		c := container.New()
		c.OverrideService("service", s)
		service, err := c.Get("service")
		assert.Nil(t, service)
		assert.EqualError(t, err, `get("service"): constructor: cannot call provider func() (interface {}, error): could not create`)
	})
	t.Run("Errors in args", func(t *testing.T) {
		s := container.NewService()
		s.SetConstructor(
			NewServer,
			container.NewDependencyProvider(func() (any, error) {
				return nil, errors.New("unexpected error")
			}),
			container.NewDependencyProvider(func() (any, error) {
				return nil, errors.New("unexpected error")
			}),
		)

		c := container.New()
		c.OverrideService("server", s)

		expected := []string{
			`get("server"): constructor args: arg #0: cannot call provider func() (interface {}, error): unexpected error`,
			`get("server"): constructor args: arg #1: cannot call provider func() (interface {}, error): unexpected error`,
		}

		svc, err := c.Get("server")
		assert.Nil(t, svc)
		assertErr.EqualErrorGroup(t, err, expected)
	})
	t.Run("OK", func(t *testing.T) {
		s := container.NewService()
		s.SetConstructor(
			NewServer,
			container.NewDependencyValue("localhost"),
			container.NewDependencyValue(8080),
		)

		c := container.New()
		c.OverrideService("server", s)

		var server *Server
		tmp, err := c.Get("server")
		require.NoError(t, copier.Copy(tmp, &server, true))
		assert.NoError(t, err)
		assert.Equal(t, "localhost", server.Host)
		assert.Equal(t, 8080, server.Port)
	})
}

func TestContainer_setServiceFields(t *testing.T) {
	t.Run("Errors", func(t *testing.T) {
		s := container.NewService()
		s.SetValue(struct{}{})
		s.SetField("Name", container.NewDependencyValue("Mary"))
		s.SetField("Age", container.NewDependencyProvider(func() (any, error) {
			return nil, errors.New("unexpected error")
		}))

		c := container.New()
		c.OverrideService("service", s)

		expected := []string{
			`get("service"): set field "Name": set (*interface {})."Name": field "Name" does not exist`,
			`get("service"): field value "Age": cannot call provider func() (interface {}, error): unexpected error`,
		}

		svc, err := c.Get("service")
		assert.Nil(t, svc)
		assertErr.EqualErrorGroup(t, err, expected)
	})
}

func TestContainer_Get_doNotCacheOnError(t *testing.T) {
	for _, tmp := range []string{"shared", "contextual", "default"} {
		scope := tmp
		t.Run(fmt.Sprintf("Scope %s", scope), func(t *testing.T) {
			counter := new(uint64)

			first := true
			fiveSvc := container.NewService()
			fiveSvc.SetConstructor(func() (any, error) {
				atomic.AddUint64(counter, 1)

				if first {
					first = false
					return nil, errors.New("my error")
				}

				return 5, nil
			})
			switch scope {
			case "shared":
				fiveSvc.SetScopeShared()
			case "contextual":
				fiveSvc.SetScopeContextual()
			case "default":
				fiveSvc.SetScopeDefault()
			}

			c := container.New()
			c.OverrideService("five", fiveSvc)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ctx = container.ContextWithContainer(ctx, c)

			five, err := c.GetInContext(ctx, "five")
			assert.EqualError(t, err, `get("five"): constructor: cannot call provider func() (interface {}, error): my error`)
			assert.Nil(t, five)

			// second invocation does not return error
			five, err = c.GetInContext(ctx, "five")
			assert.NoError(t, err)
			assert.Equal(t, 5, five)

			// third invocation should be cached
			five, err = c.GetInContext(ctx, "five")
			assert.NoError(t, err)
			assert.Equal(t, 5, five)

			// constructor has been invoked twice,
			// even tho `c.GetInContext(ctx, "five)` has been executed 3 times
			// because the result of the second invocation has been cached
			assert.Equal(t, uint64(2), atomic.LoadUint64(counter))
		})
	}
}

func TestContainer_Get_cache(t *testing.T) {
	counterCtx := new(uint64)
	counterShared := new(uint64)

	serviceCtx := container.NewService()
	serviceCtx.SetConstructor(func() any {
		atomic.AddUint64(counterCtx, 1)
		return nil
	})
	serviceCtx.SetScopeContextual()

	serviceShared := container.NewService()
	serviceShared.SetConstructor(func() any {
		atomic.AddUint64(counterShared, 1)
		return nil
	})

	c := container.New()
	c.OverrideService("serviceCtx", serviceCtx)
	c.OverrideService("serviceShared", serviceShared)

	parentCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx1 := container.ContextWithContainer(parentCtx, c)
	ctx2 := container.ContextWithContainer(parentCtx, c)
	ctx3 := container.ContextWithContainer(parentCtx, c)

	max := 100
	wg := sync.WaitGroup{}
	wg.Add(max * 4)

	for i := 0; i < max; i++ {
		go func() {
			defer wg.Done()

			_, err := c.GetInContext(ctx1, "serviceCtx")
			assert.NoError(t, err)
		}()
		go func() {
			defer wg.Done()

			_, err := c.GetInContext(ctx2, "serviceCtx")
			assert.NoError(t, err)
		}()
		go func() {
			defer wg.Done()

			_, err := c.GetInContext(ctx3, "serviceCtx")
			assert.NoError(t, err)
		}()
		go func() {
			defer wg.Done()

			_, err := c.Get("serviceShared")
			assert.NoError(t, err)
		}()
	}
	wg.Wait()

	// serviceCtx is cached 3 times, in a scope of 3 different requests
	assert.Equal(t, uint64(3), atomic.LoadUint64(counterCtx))

	// serviceShared is shared, so will be cached once, globally
	assert.Equal(t, uint64(1), atomic.LoadUint64(counterShared))
}

type Server struct {
	Host string
	Port int
}

func NewServer(host string, port int) *Server {
	return &Server{Host: host, Port: port}
}
