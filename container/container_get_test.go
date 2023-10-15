package container_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gontainer/gontainer-helpers/container"
	"github.com/gontainer/gontainer-helpers/errors"
	assertErr "github.com/gontainer/gontainer-helpers/errors/assert"
	"github.com/stretchr/testify/assert"
)

func Test_container_executeServiceCalls(t *testing.T) {
	t.Run("Errors", func(t *testing.T) {
		s := container.NewService()
		s.SetValue(struct{}{})
		s.AppendCall("SetName", container.NewDependencyProvider(func() (interface{}, error) {
			return nil, errors.New("could not fetch the name from the config")
		}))
		s.AppendCall("SetAge", container.NewDependencyValue(21))
		s.AppendCall("SetColor", container.NewDependencyValue("red"))
		s.AppendWither("WithLogger", container.NewDependencyValue(log.New(os.Stdout, "", 0)))
		// this call will be ignored, because it's after the error returned by a wither
		s.AppendCall("SetLanguage", container.NewDependencyValue("en"))

		c := container.NewContainer()
		c.OverrideService("service", s)

		expected := []string{
			"container.get(\"service\"): resolve args \"SetName\": arg #0: could not fetch the name from the config",
			"container.get(\"service\"): call \"SetAge\": invalid func `*interface {}`.\"SetAge\"",
			"container.get(\"service\"): call \"SetColor\": invalid func `*interface {}`.\"SetColor\"",
			"container.get(\"service\"): wither \"WithLogger\": invalid wither `struct {}`.\"WithLogger\"",
		}

		svc, err := c.Get("service")
		assert.Nil(t, svc)
		assertErr.EqualErrorGroup(t, err, expected)
	})
}

func Test_container_createNewService(t *testing.T) {
	t.Run("Error in provider", func(t *testing.T) {
		s := container.NewService()
		s.SetConstructor(func() (interface{}, error) {
			return nil, errors.New("could not create")
		})

		c := container.NewContainer()
		c.OverrideService("service", s)
		service, err := c.Get("service")
		assert.Nil(t, service)
		assert.EqualError(t, err, `container.get("service"): constructor: could not create`)
	})
	t.Run("Errors in args", func(t *testing.T) {
		s := container.NewService()
		s.SetConstructor(
			NewServer,
			container.NewDependencyProvider(func() (interface{}, error) {
				return nil, errors.New("unexpected error")
			}),
			container.NewDependencyProvider(func() (interface{}, error) {
				return nil, errors.New("unexpected error")
			}),
		)

		c := container.NewContainer()
		c.OverrideService("server", s)

		expected := []string{
			`container.get("server"): constructor args: arg #0: unexpected error`,
			`container.get("server"): constructor args: arg #1: unexpected error`,
		}

		svc, err := c.Get("server")
		assert.Nil(t, svc)
		assertErr.EqualErrorGroup(t, err, expected)
	})
	t.Run("Ok", func(t *testing.T) {
		s := container.NewService()
		s.SetConstructor(
			NewServer,
			container.NewDependencyValue("localhost"),
			container.NewDependencyValue(8080),
		)

		c := container.NewContainer()
		c.OverrideService("server", s)

		var server *Server
		err := c.CopyServiceTo("server", &server)
		assert.NoError(t, err)
		assert.Equal(t, "localhost", server.Host)
		assert.Equal(t, 8080, server.Port)
	})
}

func Test_container_setServiceFields(t *testing.T) {
	t.Run("Errors", func(t *testing.T) {
		s := container.NewService()
		s.SetValue(struct{}{})
		s.SetField("Name", container.NewDependencyValue("Mary"))
		s.SetField("Age", container.NewDependencyProvider(func() (interface{}, error) {
			return nil, errors.New("unexpected error")
		}))

		c := container.NewContainer()
		c.OverrideService("service", s)

		expected := []string{
			"container.get(\"service\"): set field \"Name\": set `*interface {}`.\"Name\": field `Name` does not exist",
			`container.get("service"): field value "Age": unexpected error`,
		}

		svc, err := c.Get("service")
		assert.Nil(t, svc)
		assertErr.EqualErrorGroup(t, err, expected)
	})
}

func Test_container_get_doNotCacheOnError(t *testing.T) {
	for _, tmp := range []string{"shared", "contextual", "default"} {
		scope := tmp
		t.Run(fmt.Sprintf("Scope %s", scope), func(t *testing.T) {
			firstDone := false
			fiveSvc := container.NewService()
			fiveSvc.SetConstructor(func() (interface{}, error) {
				if !firstDone {
					firstDone = true
					return nil, errors.New("my error")
				}

				return 5, nil
			})
			switch scope {
			case "shared":
				fiveSvc.ScopeShared()
			case "contextual":
				fiveSvc.ScopeContextual()
			case "default":
				fiveSvc.ScopeDefault()
			}

			c := container.NewContainer()
			c.OverrideService("five", fiveSvc)

			ctx := container.ContextWithContainer(context.Background(), c)

			five, err := c.GetWithContext(ctx, "five")
			assert.EqualError(t, err, `container.get("five"): constructor: my error`)
			assert.Nil(t, five)

			// second invocation does not return error
			five, err = c.GetWithContext(ctx, "five")
			assert.NoError(t, err)
			assert.Equal(t, 5, five)
		})
	}
}

func Test_container_get_cache(t *testing.T) {
	counterCtx := new(uint64)
	counterShared := new(uint64)

	serviceCtx := container.NewService()
	serviceCtx.SetConstructor(func() interface{} {
		atomic.AddUint64(counterCtx, 1)
		return nil
	})
	serviceCtx.ScopeContextual()

	serviceShared := container.NewService()
	serviceShared.SetConstructor(func() interface{} {
		atomic.AddUint64(counterShared, 1)
		return nil
	})

	c := container.NewContainer()
	c.OverrideService("serviceCtx", serviceCtx)
	c.OverrideService("serviceShared", serviceShared)

	ctx1 := container.ContextWithContainer(context.Background(), c)
	ctx2 := container.ContextWithContainer(context.Background(), c)

	max := 100
	wg := sync.WaitGroup{}
	wg.Add(max * 3)

	for i := 0; i < max; i++ {
		go func() {
			defer wg.Done()

			_, err := c.GetWithContext(ctx1, "serviceCtx")
			assert.NoError(t, err)
		}()
		go func() {
			defer wg.Done()

			_, err := c.GetWithContext(ctx2, "serviceCtx")
			assert.NoError(t, err)
		}()
		go func() {
			defer wg.Done()

			_, err := c.Get("serviceShared")
			assert.NoError(t, err)
		}()
	}
	wg.Wait()

	// serviceCtx is cached twice, in a scope of 2 different requests
	assert.Equal(t, uint64(2), *counterCtx)

	// serviceShared is shared, so will be cached once, globally
	assert.Equal(t, uint64(1), *counterShared)
}
