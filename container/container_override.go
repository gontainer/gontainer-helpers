// Copyright (c) 2023-2024 Bartłomiej Krukowski
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
	"fmt"
	"sync"

	"github.com/gontainer/gontainer-helpers/v3/container/internal/maps"
)

// OverrideParam adds a parameter to the [*Container].
// If a parameter with the given ID already exists, it will be replaced by the new one.
//
// See [Dependency].
func (c *Container) OverrideParam(paramID string, d Dependency) {
	c.globalLocker.Lock()
	defer c.globalLocker.Unlock()

	overrideParam(c, paramID, d)
}

func (c *Container) OverrideParams(params map[string]Dependency) {
	c.globalLocker.Lock()
	defer c.globalLocker.Unlock()

	for _, id := range maps.SortedStringKeys(params) {
		overrideParam(c, id, params[id])
	}
}

// OverrideService adds a service to the [*Container].
// If a service with the given ID already exists, it will be replaced by the new one.
//
// See [Service].
func (c *Container) OverrideService(serviceID string, s Service) {
	c.globalLocker.Lock()
	defer c.globalLocker.Unlock()

	overrideService(c, serviceID, s)
}

func (c *Container) OverrideServices(services map[string]Service) {
	c.globalLocker.Lock()
	defer c.globalLocker.Unlock()

	for _, id := range maps.SortedStringKeys(services) {
		overrideService(c, id, services[id])
	}
}

func overrideService(c *Container, serviceID string, s Service) {
	c.invalidateGraph()

	switch s.scope {
	case
		scopeDefault,
		scopeShared,
		scopeContextual,
		scopeNonShared:
	default:
		panic(fmt.Sprintf("overrideService(%+q): invalid scope %+q", serviceID, s.scope.String()))
	}

	if !s.hasCreationMethod {
		panic(fmt.Sprintf("overrideService(%+q): service has neither a constructor nor a factory nor a value", serviceID))
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
	c.invalidateGraph()

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
