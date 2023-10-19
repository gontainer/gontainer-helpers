package groupcontext_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gontainer/gontainer-helpers/container/internal/groupcontext"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("Context already cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		done := false

		g := groupcontext.New()
		g.Add(ctx)

		go func() {
			g.Wait()
			done = true
		}()
		time.Sleep(time.Millisecond * 10)

		assert.True(t, done)
	})

	t.Run("Context never cancelled", func(t *testing.T) {
		ctxs := []context.Context{
			context.Background(),
			context.WithValue(context.Background(), "my key", "my value"), //nolint:staticcheck
		}
		for i, ctx := range ctxs {
			ctx := ctx
			t.Run(fmt.Sprintf("#%d", i), func(t *testing.T) {
				done := false

				g := groupcontext.New()
				g.Add(ctx)

				go func() {
					g.Wait()
					done = false
				}()
				time.Sleep(time.Millisecond * 10)

				assert.False(t, done)
			})
		}
	})

	t.Run("Wait till context is done", func(t *testing.T) {
		ctx1, cancel1 := context.WithCancel(context.Background())
		ctx2, cancel2 := context.WithCancel(context.Background())
		childCtx2 := context.WithValue(ctx2, "my-key", "my-value") //nolint:staticcheck

		g := groupcontext.New()
		g.Add(ctx1)
		g.Add(childCtx2)

		go func() {
			time.Sleep(time.Millisecond * 100)
			cancel1()
		}()

		go func() {
			time.Sleep(time.Millisecond * 200)
			cancel2() // it cancels the child context as well
		}()

		s := time.Now()
		g.Wait()
		assert.GreaterOrEqual(t, time.Since(s), time.Millisecond*200)
	})
}
