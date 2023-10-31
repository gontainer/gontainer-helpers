package container

// DecoratorPayload is the very first argument passed to every decorator always.
//
// See [*Container.AddDecorator].
type DecoratorPayload struct {
	Tag       string
	ServiceID string
	Service   any
}

// AddDecorator adds decorator for the given tag.
// Decorator is a special function that can decorate all services tagged by the given tag.
//
// See [DecoratorPayload].
func (c *Container) AddDecorator(tag string, decorator any, deps ...Dependency) {
	c.globalLocker.Lock()
	defer c.globalLocker.Unlock()

	c.invalidateGraph()

	c.decorators = append(c.decorators, serviceDecorator{
		tag:  tag,
		fn:   decorator,
		deps: deps,
	})
}
