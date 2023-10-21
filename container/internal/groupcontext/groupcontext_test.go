package groupcontext_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gontainer/gontainer-helpers/container/internal/groupcontext"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("Invalid context", func(t *testing.T) {
		defer func() {
			assert.Equal(t, "a receive from a nil channel blocks forever: ctx.Done() == nil", recover())
		}()

		g := groupcontext.New()
		g.Add(context.TODO())
	})
	t.Run("Context already cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		counter := new(int64)

		g := groupcontext.New()
		g.Add(ctx)

		go func() {
			g.Wait()
			atomic.AddInt64(counter, 1)
		}()
		time.Sleep(time.Millisecond * 100)

		assert.Equal(t, int64(1), atomic.LoadInt64(counter))
	})

	t.Run("Context never cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel() // it's called after executing the tests
		ctxs := []context.Context{
			ctx,
			context.WithValue(ctx, "my key", "my value"), //nolint:staticcheck
		}
		for i, ctx := range ctxs {
			ctx := ctx
			t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
				counter := new(int64)

				g := groupcontext.New()
				g.Add(ctx)

				go func() {
					g.Wait()
					atomic.AddInt64(counter, 1)
				}()
				time.Sleep(time.Millisecond * 100)

				assert.Equal(t, int64(0), atomic.LoadInt64(counter))
			})
		}
	})

	t.Run("Wait till context is done", func(t *testing.T) {
		ctx1, cancel1 := context.WithCancel(context.Background())
		ctx2, cancel2 := context.WithCancel(context.Background())
		childCtx2 := context.WithValue(ctx2, "my-key", "my-value") //nolint:staticcheck

		counter := new(int64)

		g := groupcontext.New()
		g.Add(ctx1)
		g.Add(childCtx2)

		go func() {
			time.Sleep(time.Millisecond * 100)
			atomic.AddInt64(counter, 1)
			cancel1()
		}()

		go func() {
			time.Sleep(time.Millisecond * 200)
			atomic.AddInt64(counter, 1)
			cancel2() // it cancels the child context as well
		}()

		s := time.Now()
		g.Wait()
		assert.GreaterOrEqual(t, time.Since(s), time.Millisecond*200)
		assert.Equal(t, int64(2), atomic.LoadInt64(counter))
	})
}
