package hotswap_test

import (
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gontainer/gontainer-helpers/v2/container"
	"github.com/gontainer/gontainer-helpers/v2/container/internal/examples/hotswap"
	"github.com/stretchr/testify/require"
)

func TestContainer_Server(t *testing.T) {
	const max = 50

	c := hotswap.NewContainer()

	s := httptest.NewServer(c.ServeMux())
	defer s.Close()

	wg := sync.WaitGroup{}

	client := s.Client()

	for i := 0; i < max; i++ {
		i := i
		wg.Add(2)
		go func() {
			defer wg.Done()

			r, err := client.Get(s.URL + "/")
			require.NoError(t, err)

			buff, err := ioutil.ReadAll(r.Body)
			require.NoError(t, err)

			vals := strings.SplitN(string(buff), "=", 2)

			require.Equal(t, vals[0], vals[1])
		}()
		go func() {
			defer wg.Done()

			// the commented code does not guarantee atomicity
			// we have to use HotSwap
			// c.OverrideParam("a", container.NewDependencyValue(i))
			// time.Sleep(time.Millisecond * 2)
			// c.OverrideParam("b", container.NewDependencyValue(i))

			// HotSwap guarantees atomicity
			c.HotSwap(func(c container.MutableContainer) {
				c.OverrideParam("a", container.NewDependencyValue(i))
				// sleep to simulate an edge case
				time.Sleep(time.Millisecond * 2)
				c.OverrideParam("b", container.NewDependencyValue(i))
			})
		}()
	}

	wg.Wait()
}
