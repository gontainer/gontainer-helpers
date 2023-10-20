package container

import (
	"sync"
)

func (c *container) OverrideParam(paramID string, d Dependency) {
	c.globalLocker.Lock()
	defer c.globalLocker.Unlock()

	c.graphBuilder.invalidate()

	c.paramContainer.OverrideParam(paramID, d)
}

func (c *container) OverrideService(serviceID string, s Service) {
	c.globalLocker.Lock()
	defer c.globalLocker.Unlock()

	// TODO: worth to consider whether we should wait till all contexts are done, e.g.:
	//	c.HotSwap(func(m MutableContainer) {
	//		m.OverrideService(serviceID, s)
	//	})

	overrideService(c, serviceID, s)
}

func overrideService(c *container, serviceID string, s Service) {
	c.graphBuilder.invalidate()

	switch s.scope {
	case
		scopeDefault,
		scopeShared,
		scopeContextual,
		scopeNonShared:
	default:
		panic("TODO")
	}

	c.services[serviceID] = s
	c.cacheShared.delete(serviceID)
	switch s.scope {
	case
		scopeDefault,
		scopeShared,
		scopeContextual:
		if _, ok := c.serviceLockers[serviceID]; !ok {
			c.serviceLockers[serviceID] = &sync.Mutex{}
		}
	default:
		delete(c.serviceLockers, serviceID)
	}
}
