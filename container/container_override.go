package container

import (
	"fmt"
	"sync"
)

func (c *container) OverrideParam(paramID string, d Dependency) {
	c.globalLocker.Lock()
	defer c.globalLocker.Unlock()

	c.graphBuilder.invalidate()

	switch d.type_ {
	case
		dependencyValue,
		dependencyParam,
		dependencyProvider:
	default:
		panic(fmt.Sprintf("container.OverrideParam does not accept `%s`", d.type_.String()))
	}

	c.params[paramID] = d
	c.cacheParams.delete(paramID)
	c.paramsLockers[paramID] = &sync.Mutex{}
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
	c.cacheSharedServices.delete(serviceID)
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
