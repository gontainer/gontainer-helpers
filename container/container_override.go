package container

import (
	"sync"
)

func overrideService(c *container, serviceID string, s Service) {
	c.graphBuilder.invalidate()

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
