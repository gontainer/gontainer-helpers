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
ContextWithContainer creates a new context, and attaches the given [Container] to it.
The given context MUST be cancellable (ctx.Done() != nil).

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

	ctx := context.WithValue(parent, c.id, newSafeMap())
	c.groupContext.Add(ctx)
	return ctx
}
