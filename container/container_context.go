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

type contextableContainer interface {
	getContainerID() ctxKey
	getGroupContext() interface{ Add(context.Context) }
	getContextLocker() rwlocker
}

func ContextWithContainer(parent context.Context, container contextableContainer) context.Context {
	container.getContextLocker().RLock()
	defer container.getContextLocker().RUnlock()

	ctx := context.WithValue(parent, container.getContainerID(), newSafeMap())
	container.getGroupContext().Add(ctx)
	return ctx
}
