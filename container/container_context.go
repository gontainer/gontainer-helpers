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

type contextableContainer interface {
	getContainerID() ctxKey
	getGroupContext() interface{ Add(context.Context) }
}

func ContextWithContainer(parent context.Context, container contextableContainer) context.Context {
	ctx := context.WithValue(parent, container.getContainerID(), newSafeMap())
	container.getGroupContext().Add(ctx)
	return ctx
}
