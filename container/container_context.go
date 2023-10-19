package container

import (
	"context"
)

func (c *container) getContainerID() ctxKey {
	return c.id
}

func (c *container) getGroupContext() interface{ Add(context.Context) } {
	return c.groupContext
}

func (c *container) getContextLocker() rwlocker {
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
