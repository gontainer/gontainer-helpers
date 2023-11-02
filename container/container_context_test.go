// Copyright (c) 2023 Bartłomiej Krukowski
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
	"testing"

	"github.com/gontainer/gontainer-helpers/v3/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type myWrappedContainer struct {
	*container.Container
}

func newMyWrappedContainer() *myWrappedContainer {
	return &myWrappedContainer{
		Container: container.New(),
	}
}

func TestContextWithContainer(t *testing.T) {
	t.Run("Nested context", func(t *testing.T) {
		c := container.New()
		s := container.NewService()
		s.SetConstructor(func() any {
			return new(int64)
		})
		s.SetScopeContextual()
		c.OverrideService("ptr", s)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctx = container.ContextWithContainer(ctx, c)

		ctxNested := container.ContextWithContainer(ctx, c)

		// the following line creates a context with some dummy values,
		// because we want ctxNested != ctxNested2
		type dummyKey string
		ctxNested2 := context.WithValue(ctxNested, dummyKey("key"), "dummy value")
		ctxNested2 = container.ContextWithContainer(ctxNested2, c)

		ctxAnother, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctxAnother = container.ContextWithContainer(ctxAnother, c)

		ptr1, err1 := c.GetInContext(ctx, "ptr")
		ptr2, err2 := c.GetInContext(ctxNested, "ptr")
		ptr3, err3 := c.GetInContext(ctxNested2, "ptr")
		ptr4, err4 := c.GetInContext(ctxAnother, "ptr")

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.NoError(t, err3)
		require.NoError(t, err4)

		assert.Same(t, ptr1, ptr2, "Nested context should inherit contextBag from the parent")
		assert.Same(t, ptr1, ptr3, "Nested context should inherit contextBag from the parent")
		assert.NotSame(t, ptr1, ptr4)
	})
	t.Run("Wrapped *Container", func(t *testing.T) {
		// Make sure that approach (embedding struct implementing an interface with unexported methods)
		// works in all GO's versions
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		container.ContextWithContainer(ctx, newMyWrappedContainer())
	})
	t.Run("Invalid input", func(t *testing.T) {
		t.Run("ctx.Done() == nil", func(t *testing.T) {
			defer func() {
				assert.Equal(t, `ctx.Done() == nil: a receive from a nil channel blocks forever`, recover())
			}()
			container.ContextWithContainer(context.Background(), container.New())
		})
		t.Run("container == nil", func(t *testing.T) {
			defer func() {
				assert.Equal(t, `nil container`, recover())
			}()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			container.ContextWithContainer(ctx, nil)
		})
		t.Run("context == nil", func(t *testing.T) {
			defer func() {
				assert.Equal(t, `nil context`, recover())
			}()
			container.ContextWithContainer(nil, nil) //nolint:staticcheck
		})
	})
}
