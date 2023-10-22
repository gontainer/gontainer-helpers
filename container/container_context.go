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

func (c *Container) getContainerID() ctxKey {
	return c.id
}

func (c *Container) getGroupContext() interface{ Add(context.Context) } {
	return c.groupContext
}

func (c *Container) getContextLocker() rwlocker {
	return c.contextLocker
}

type contextualContainer interface {
	getContainerID() ctxKey
	getGroupContext() interface{ Add(context.Context) }
	getContextLocker() rwlocker
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
func ContextWithContainer(parent context.Context, container contextualContainer) context.Context {
	container.getContextLocker().RLock()
	defer container.getContextLocker().RUnlock()

	ctx := context.WithValue(parent, container.getContainerID(), newSafeMap())
	container.getGroupContext().Add(ctx)
	return ctx
}
