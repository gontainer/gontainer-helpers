package container

import (
	"context"
)

func (c *container) getContainerID() ctxKey {
	return c.id
}

func ContextWithContainer(parent context.Context, container interface{ getContainerID() ctxKey }) context.Context {
	return context.WithValue(parent, container.getContainerID(), make(map[string]interface{}))
}
