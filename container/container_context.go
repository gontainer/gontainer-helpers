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

Deprecated: do not use it.
*/
func (c *Container) Self() *Container {
	return c
}

/*
ContextWithContainer creates a new context, and attaches the given [Container] to it.
The given context MUST be cancellable (ctx.Done() != nil).
Using the interface instead of the [*Container] lets us for using the struct embedding.

# Example

	c := container.New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = container.ContextWithContainer(ctx, c)

# HTTP handler

	var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := container.ContextWithContainer(r.Context(), c)
		r = r.Clone(ctx)

		// your code
	})
*/
func ContextWithContainer(parent context.Context, container Self) context.Context {
	c := container.Self()

	c.contextLocker.RLock()
	defer c.contextLocker.RUnlock()

	ctx := context.WithValue(parent, c.id, newSafeMap())
	c.groupContext.Add(ctx)
	return ctx
}
