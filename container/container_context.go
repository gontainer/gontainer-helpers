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

// Self has been designed for the struct embedding and compatibility with the func [ContextWithContainer].
//
// See [*Container.Self].
type Self interface {
	Self() *Container
}

/*
Self returns itself.
It is designed for the struct embedding and compatibility with the func [ContextWithContainer].

	type MyContainer struct {
		*container.Self
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
func (c *Container) Self() *Container {
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
 1. [*Container.Self]
 2. [HTTPHandlerWithContainer]
*/
func ContextWithContainer(parent context.Context, container Self) context.Context {
	c := container.Self()

	c.contextLocker.RLock()
	defer c.contextLocker.RUnlock()

	ctx := context.WithValue(parent, c.id, newSafeMap())
	c.groupContext.Add(ctx)
	return ctx
}
