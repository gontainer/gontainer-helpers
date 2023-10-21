package container

import (
	"fmt"
	"sync"
)

func (c *Container) OverrideParam(paramID string, d Dependency) {
	c.globalLocker.Lock()
	defer c.globalLocker.Unlock()

	overrideParam(c, paramID, d)
}

func (c *Container) OverrideService(serviceID string, s Service) {
	c.globalLocker.Lock()
	defer c.globalLocker.Unlock()

	// TODO: worth to consider whether we should wait till all contexts are done, e.g.:
	//	c.HotSwap(func(m MutableContainer) {
	//		m.OverrideService(serviceID, s)
	//	})

	overrideService(c, serviceID, s)
}

func overrideService(c *Container, serviceID string, s Service) {
	c.graphBuilder.invalidate()

	switch s.scope {
	case
		scopeDefault,
		scopeShared,
		scopeContextual,
		scopeNonShared:
	default:
		panic(fmt.Sprintf("overrideService: invalid scope %s", s.scope.String()))
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

func overrideParam(c *Container, paramID string, d Dependency) {
	c.graphBuilder.invalidate()

	switch d.type_ {
	case
		dependencyValue,
		dependencyParam,
		dependencyProvider:
	default:
		panic(fmt.Sprintf("overrideParam: invalid dependency: %s", d.type_.String()))
	}

	c.params[paramID] = d
	c.cacheParams.delete(paramID)
	c.paramsLockers[paramID] = &sync.Mutex{}
}
