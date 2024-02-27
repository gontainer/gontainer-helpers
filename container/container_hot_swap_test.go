// Copyright (c) 2023-2024 Bartłomiej Krukowski
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is furnished
// to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package container_test

import (
	"context"
	"io/ioutil"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gontainer/gontainer-helpers/v3/container"
	"github.com/gontainer/gontainer-helpers/v3/container/internal/examples/hotswap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainer_HotSwap(t *testing.T) {
	t.Run("Invalidate cache", func(t *testing.T) {
		serviceCounter := 0
		paramCounter := 0

		s := container.NewService()
		s.SetConstructor(func() any {
			serviceCounter++
			return nil
		})

		c := container.New()
		c.OverrideService("service", s)
		c.OverrideParam("param", container.NewDependencyProvider(func() any {
			paramCounter++
			return nil
		}))

		for i := 0; i < 10; i++ {
			_, _ = c.Get("service")
			_, _ = c.GetParam("param")
		}

		// values are cached, so providers should be invoked once only
		assert.Equal(t, 1, serviceCounter)
		assert.Equal(t, 1, paramCounter)

		c.HotSwap(func(c container.MutableContainer) {
			c.InvalidateAllServicesCache()
			c.InvalidateAllParamsCache()
		})

		for i := 0; i < 10; i++ {
			_, _ = c.Get("service")
			_, _ = c.GetParam("param")
		}

		// the cache has been invalidated, so providers should be invoked again
		assert.Equal(t, 2, serviceCounter)
		assert.Equal(t, 2, paramCounter)
	})
	t.Run("Override", func(t *testing.T) {
		s := container.NewService()
		s.SetValue("old service")

		c := container.New()
		c.OverrideService("service", s)
		c.OverrideParam("param", container.NewDependencyValue("old param"))

		svc, _ := c.Get("service")
		param, _ := c.GetParam("param")

		assert.Equal(t, "old service", svc)
		assert.Equal(t, "old param", param)

		c.HotSwap(func(c container.MutableContainer) {
			s := container.NewService()
			s.SetValue("new service")

			c.OverrideService("service", s)
			c.OverrideParam("param", container.NewDependencyValue("new param"))
		})

		svc, _ = c.Get("service")
		param, _ = c.GetParam("param")

		assert.Equal(t, "new service", svc)
		assert.Equal(t, "new param", param)
	})
	t.Run("Wait for <-ctx.Done()", func(t *testing.T) {
		c := container.New()
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

		names := map[uint64]string{
			1: "Jane",
			2: "John",
		}

		svcPerson := container.NewService()
		svcPerson.SetConstructor(func() Person {
			return Person{
				Name: names[atomic.AddUint64(counter, 1)],
			}
		})
		svcPerson.Tag("person", 0)

		svcPeople := container.NewService()
		svcPeople.SetConstructor(
			func(ppl []Person) People {
				return People{
					People: ppl,
				}
			},
			container.NewDependencyTag("person"),
		)

		c := container.New()
		c.OverrideService("person", svcPerson)
		c.OverrideService("people", svcPeople)

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

					tmp, err := c.GetInContext(ctx, "person")
					require.NoError(t, err)
					p := tmp.(Person)

					tmp, err = c.GetInContext(ctx, "people")
					require.NoError(t, err)
					ppl := tmp.(People)

					assert.Equal(t, ppl.People[0].Name, p.Name)
				}()
			}
		}

		runGoroutines()
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.HotSwap(func(mc container.MutableContainer) {
				mc.InvalidateServicesCache("person", "people")
			})
		}()
		runGoroutines()
		wg.Wait()

		assert.Equal(t, uint64(2), atomic.LoadUint64(counter))
	})
	t.Run("Server", func(t *testing.T) {
		const (
			max   = 100
			delay = time.Millisecond * 3
		)

		c := hotswap.NewContainer()

		s := httptest.NewServer(c.HTTPHandler())
		defer s.Close()

		client := s.Client()

		performTest := func(hotSwap bool) (consistent bool) {
			inconsistency := uint64(0)
			wg := sync.WaitGroup{}
			for i := 0; i < max; i++ {
				i := i
				wg.Add(2)

				// Perform an HTTP-request and check whether it returns "c.ParamA() == c.ParamB()"
				go func() {
					defer wg.Done()

					r, err := client.Get(s.URL)
					require.NoError(t, err)

					buff, err := ioutil.ReadAll(r.Body)
					require.NoError(t, err)

					if string(buff) != "p.ParamA() == p.ParamB()" {
						atomic.AddUint64(&inconsistency, 1)
					}
				}()

				// Override params "a" and "b" in the container
				go func() {
					defer wg.Done()

					if hotSwap { // HotSwap guarantees atomicity
						c.HotSwap(func(c container.MutableContainer) {
							c.OverrideParam("a", container.NewDependencyValue(i))
							// sleep to simulate an edge case
							time.Sleep(delay)
							c.OverrideParam("b", container.NewDependencyValue(i))
						})
						return
					}

					// Changing the following params is not an atomic operation.
					// It is possible that another goroutines read a new value of "a", and an old value of "b".
					c.OverrideParam("a", container.NewDependencyValue(i))
					time.Sleep(delay)
					c.OverrideParam("b", container.NewDependencyValue(i))
				}()
			}
			wg.Wait()
			return inconsistency == 0
		}

		assert.True(t, performTest(true), "Expected consistent results")
		// perform it twice in case of consistent results in ci
		assert.False(t, performTest(false) && performTest(false), "Expected inconsistent results") //nolint:staticcheck
	})
}
