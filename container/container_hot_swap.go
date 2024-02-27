// Copyright (c) 2023–present Bartłomiej Krukowski
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is furnished
// to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package container

import (
	"sync"

	"github.com/gontainer/gontainer-helpers/v3/container/internal/maps"
)

// MutableContainer represents the interface that is required by [*Container.HotSwap].
type MutableContainer interface {
	OverrideService(serviceID string, s Service)
	OverrideServices(services map[string]Service)
	OverrideParam(paramID string, d Dependency)
	OverrideParams(params map[string]Dependency)
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

func (m *mutableContainer) OverrideService(serviceID string, s Service) {
	m.locker.Lock()
	defer m.locker.Unlock()

	overrideService(m.parent, serviceID, s)
}

func (m *mutableContainer) OverrideServices(services map[string]Service) {
	m.locker.Lock()
	defer m.locker.Unlock()

	for _, id := range maps.SortedStringKeys(services) {
		overrideService(m.parent, id, services[id])
	}
}

func (m *mutableContainer) OverrideParam(paramID string, d Dependency) {
	m.locker.Lock()
	defer m.locker.Unlock()

	overrideParam(m.parent, paramID, d)
}

func (m *mutableContainer) OverrideParams(params map[string]Dependency) {
	m.locker.Lock()
	defer m.locker.Unlock()

	for _, id := range maps.SortedStringKeys(params) {
		overrideParam(m.parent, id, params[id])
	}
}

func (m *mutableContainer) InvalidateServicesCache(servicesIDs ...string) {
	m.locker.Lock()
	defer m.locker.Unlock()

	for _, sID := range servicesIDs {
		m.parent.cacheSharedServices.delete(sID)
	}
}

func (m *mutableContainer) InvalidateAllServicesCache() {
	for sID := range m.parent.services {
		m.InvalidateServicesCache(sID)
	}
}

func (m *mutableContainer) InvalidateParamsCache(paramsIDs ...string) {
	m.locker.Lock()
	defer m.locker.Unlock()

	for _, pID := range paramsIDs {
		m.parent.cacheParams.delete(pID)
	}
}

func (m *mutableContainer) InvalidateAllParamsCache() {
	for pID := range m.parent.params {
		m.InvalidateParamsCache(pID)
	}
}

/*
HotSwap lets safely modify the given [*Container] in a concurrent environment.
It waits till all contexts are done, then locks the container till the passed function is executed.

	c.HotSwap(func (c container.MutableContainer) {
		c.OverrideParam("db.password", dependency.Value("new-password"))
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
