package container

import "sync"

type MutableContainer interface {
	OverrideService(serviceID string, s Service)
	OverrideParam(paramID string, d Dependency)
	InvalidateServicesCache(servicesIDs ...string)
	InvalidateAllServicesCache()
	InvalidateParamsCache(paramsIDs ...string)
	InvalidateAllParamsCache()
}

type mutableContainer struct {
	parent *Container
	locker sync.Locker
}

func newMutableContainer(parent *Container) *mutableContainer {
	return &mutableContainer{parent: parent, locker: &sync.Mutex{}}
}

func (m mutableContainer) OverrideService(serviceID string, s Service) {
	m.locker.Lock()
	defer m.locker.Unlock()

	overrideService(m.parent, serviceID, s)
}

func (m mutableContainer) OverrideParam(paramID string, d Dependency) {
	m.locker.Lock()
	defer m.locker.Unlock()

	overrideParam(m.parent, paramID, d)
}

func (m mutableContainer) InvalidateServicesCache(servicesIDs ...string) {
	m.locker.Lock()
	defer m.locker.Unlock()

	for _, sID := range servicesIDs {
		m.parent.cacheSharedServices.delete(sID)
	}
}

func (m mutableContainer) InvalidateAllServicesCache() {
	for sID := range m.parent.services {
		m.InvalidateServicesCache(sID)
	}
}

func (m mutableContainer) InvalidateParamsCache(paramsIDs ...string) {
	m.locker.Lock()
	defer m.locker.Unlock()

	for _, pID := range paramsIDs {
		m.parent.cacheParams.delete(pID)
	}
}

func (m mutableContainer) InvalidateAllParamsCache() {
	for pID := range m.parent.params {
		m.InvalidateParamsCache(pID)
	}
}

/*
HotSwap lets safely modify the given [Container] in a concurrent environment.
It waits for all `<-ctx.Done()`, then locks all invocations of [ContextWithContainer] for the same [Container].

	c.HotSwap(func (c container.MutableContainer) {
		c.OverrideParam("db.password", container.NewDependencyValue("new-password"))
	})
*/
func (c *Container) HotSwap(fn func(MutableContainer)) {
	// lock the executions of ContextWithContainer
	c.contextLocker.Lock()
	defer c.contextLocker.Unlock()

	// wait till all contexts are done
	c.groupContext.Wait()

	// lock all operations on the Container
	c.globalLocker.Lock()
	defer c.globalLocker.Unlock()

	defer c.graphBuilder.warmUp()

	fn(newMutableContainer(c))
}
