package container_test

import (
	"context"
	"io/ioutil"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gontainer/gontainer-helpers/v2/container"
	"github.com/gontainer/gontainer-helpers/v2/container/internal/examples/hotswap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainer_HotSwap(t *testing.T) {
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

					if string(buff) != "c.ParamA() == c.ParamB()" {
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
