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

package container

import (
	"context"
)

func (c *Container) contextBag(ctx context.Context) keyValue {
	bag := ctx.Value(c.id)
	if bag == nil {
		panic("the given context is not attached to the given container, call `ctx = container.ContextWithContainer(ctx, c)`")
	}
	return bag.(keyValue)
}

// Root has been designed for the struct embedding and compatibility with the func [ContextWithContainer].
//
// See [*Container.Root].
type Root interface {
	Root() *Container
}

/*
Root returns itself.
It is designed for the struct embedding and compatibility with the func [ContextWithContainer].

	type MyContainer struct {
		*container.Container
	}

	func (c *MyContainer) Server() *http.Server {
		s, err := c.Get("server")
		if err != nil {
			panic(err)
		}
		return s.(*http.Server)
	}

	var c *MyContainer // build container here
	// some code
	ctx = container.ContextWithContainer(ctx, c) // it works, even tho c is of the type *MyContainer

Deprecated: do not use it, it has been designed for the internal purposes only.
*/
func (c *Container) Root() *Container {
	return c
}

/*
ContextWithContainer creates a new context, and attaches the given container to it.
The given context MUST be cancellable (ctx.Done() != nil).
Till the given context is not cancelled, all invocations of HotSwap stuck.

If the `parent` context is already attached to the given container, it returns the original `parent` context.

# Example

	c := container.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = container.ContextWithContainer(ctx, c)

# HTTP handler

	var (
		h http.Handler
		c *container.Container
	)
	h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// your code
	})
	c = container.HTTPHandlerWithContainer(h, c)

See:
  - [*Container.HotSwap]
  - [*Container.Root]
  - [HTTPHandlerWithContainer]
*/
func ContextWithContainer(parent context.Context, container Root) context.Context {
	if parent == nil {
		panic("nil context")
	}

	if container == nil {
		panic("nil container")
	}

	c := container.Root()

	c.contextLocker.RLock()
	defer c.contextLocker.RUnlock()

	if parent.Value(c.id) != nil {
		return parent
	}

	ctx := context.WithValue(parent, c.id, newSafeMap())
	c.groupContext.Add(ctx)
	return ctx
}
