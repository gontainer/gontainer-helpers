package hotswap_test

import (
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gontainer/gontainer-helpers/v2/container"
	"github.com/gontainer/gontainer-helpers/v2/container/internal/examples/hotswap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainer_Server(t *testing.T) {
	const max = 50

	c := hotswap.NewContainer()

	s := httptest.NewServer(c.ServeMux())
	defer s.Close()

	client := s.Client()

	runGoroutines := func(hotSwap bool) (consistent bool) {
		inconsistency := uint64(0)
		wg := sync.WaitGroup{}
		for i := 0; i < max; i++ {
			i := i
			wg.Add(2)
			go func() {
				defer wg.Done()

				r, err := client.Get(s.URL)
				require.NoError(t, err)

				buff, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)

				vals := strings.SplitN(string(buff), "=", 2)

				if vals[0] != vals[1] {
					atomic.AddUint64(&inconsistency, 1)
				}
			}()
			go func() {
				defer wg.Done()

				if hotSwap { // HotSwap guarantees atomicity
					c.HotSwap(func(c container.MutableContainer) {
						c.OverrideParam("a", container.NewDependencyValue(i))
						// sleep to simulate an edge case
						time.Sleep(time.Millisecond * 2)
						c.OverrideParam("b", container.NewDependencyValue(i))
					})
					return
				}

				// Changing the following params is not an atomic operation.
				// It is possible that another goroutines read a new value of "a", and a new value of "b".
				c.OverrideParam("a", container.NewDependencyValue(i))
				time.Sleep(time.Millisecond * 2)
				c.OverrideParam("b", container.NewDependencyValue(i))
			}()
		}
		wg.Wait()
		return inconsistency == 0
	}

	assert.True(t, runGoroutines(true))
	assert.False(t, runGoroutines(false))
}
