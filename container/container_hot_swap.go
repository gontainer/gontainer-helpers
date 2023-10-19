package container

type MutableContainer interface {
	OverrideService(serviceID string, s Service)
	InvalidateServiceCache(serviceIDs ...string) // TODO naming
	InvalidateAllServicesCache()                 // TODO naming
}

type mutableContainer struct {
	parent *container
}

func (m mutableContainer) OverrideService(serviceID string, s Service) {
	overrideService(m.parent, serviceID, s)
}

func (m mutableContainer) InvalidateServiceCache(serviceIDs ...string) {
	for _, sID := range serviceIDs {
		m.parent.cacheShared.delete(sID)
	}
}

func (m mutableContainer) InvalidateAllServicesCache() {
	for sID := range m.parent.services {
		m.InvalidateServiceCache(sID)
	}
}

// TODO
func (c *container) HotSwap(fn func(MutableContainer)) {
	// lock the executions of ContextWithContainer
	c.contextLocker.Lock()
	defer c.contextLocker.Unlock()

	// wait till all contexts are done
	c.groupContext.Wait()

	// lock all operations on the container
	c.globalLocker.Lock()
	defer c.globalLocker.Unlock()

	defer c.graphBuilder.warmUp()

	fn(mutableContainer{parent: c})
}
