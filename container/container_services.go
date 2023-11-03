// Copyright (c) 2023 Bartłomiej Krukowski
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
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/gontainer/gontainer-helpers/v3/caller"
	"github.com/gontainer/gontainer-helpers/v3/grouperror"
	"github.com/gontainer/gontainer-helpers/v3/setter"
)

func contextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// Get returns a service with the given ID.
func (c *Container) Get(serviceID string) (any, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	c.warmUpGraph()

	return c.get(serviceID, newSafeMap())
}

// GetInContext returns a service with the given ID.
// It returns an error if the context is done.
func (c *Container) GetInContext(ctx context.Context, id string) (any, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	c.warmUpGraph()

	// contextBag checks whether the context is valid,
	// so it must be executed before checking whether the context is done
	bag := c.contextBag(ctx)
	if contextDone(ctx) {
		return nil, fmt.Errorf("GetInContext(%+q): ctx.Done() closed: %w", id, ctx.Err())
	}

	return c.get(id, bag)
}

// GetTaggedBy returns all services tagged by the given tag.
// The order is determined by the priority (descending) and service ID (ascending).
//
// See [Service.Tag].
func (c *Container) GetTaggedBy(tag string) ([]any, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	c.warmUpGraph()

	return c.getTaggedBy(tag, newSafeMap())
}

// GetTaggedByInContext returns all services tagged by the given tag.
// It returns an error if the context is done.
//
// See [Container.GetTaggedBy].
func (c *Container) GetTaggedByInContext(ctx context.Context, tag string) ([]any, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	c.warmUpGraph()

	// contextBag checks whether the context is valid,
	// so it must be executed before checking whether the context is done
	bag := c.contextBag(ctx)
	if contextDone(ctx) {
		return nil, fmt.Errorf("GetTaggedByInContext(%+q): ctx.Done() closed: %w", tag, ctx.Err())
	}

	return c.getTaggedBy(tag, bag)
}

// IsTaggedBy returns true whenever the given service is tagged by the given tag.
func (c *Container) IsTaggedBy(serviceID string, tag string) bool {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	s, exists := c.services[serviceID]
	if !exists {
		return false
	}
	_, ok := s.tags[tag]
	return ok
}

func (c *Container) get(id string, contextualBag keyValue) (result any, err error) {
	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("get(%+q): ", id), err)
		}
	}()

	svc, ok := c.services[id]
	if !ok {
		return nil, errors.New("service does not exist")
	}

	currentScope := svc.scope
	if currentScope == scopeDefault {
		currentScope = c.graphBuilder.resolveScope(id)
	}
	switch currentScope { // do not create cached objects more than once in concurrent invocations
	case
		scopeShared,
		scopeContextual:

		var cache keyValue
		switch currentScope {
		case scopeShared:
			cache = c.cacheSharedServices
		case scopeContextual:
			cache = contextualBag
		}

		c.serviceLockers[id].Lock()
		defer c.serviceLockers[id].Unlock()

		if s, cached := cache.get(id); cached {
			return s, nil
		}

		defer func() {
			if err == nil { // do not cache on error
				cache.set(id, result)
			}
		}()
	}

	err = c.graphBuilder.serviceCircularDeps(id)
	if err != nil {
		return nil, grouperror.Prefix("circular dependencies: ", err)
	}

	// constructor
	result, err = c.createNewService(svc, contextualBag)
	if err != nil {
		return nil, err
	}

	// fields
	result, err = c.setServiceFields(result, svc, contextualBag)
	if err != nil {
		return nil, err
	}

	// calls
	result, err = c.executeServiceCalls(result, svc, contextualBag)
	if err != nil {
		return nil, err
	}

	// decorators
	result, err = c.decorateService(id, result, svc, contextualBag)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Container) createNewService(svc Service, contextualBag keyValue) (any, error) {
	result := svc.value

	if svc.constructor != nil {
		params, err := c.resolveDeps(contextualBag, svc.constructorDeps...)
		if err != nil {
			return nil, grouperror.Prefix("constructor args: ", err)
		}
		result, err = caller.CallProvider(svc.constructor, params, convertArgs)
		if err != nil {
			return nil, grouperror.Prefix("constructor: ", err)
		}
	}

	return result, nil
}

func (c *Container) setServiceFields(
	result any,
	svc Service,
	contextualBag keyValue,
) (any, error) {
	var errs []error
	for _, f := range svc.fields {
		fieldVal, err := c.resolveDep(contextualBag, f.dep)
		if err != nil {
			errs = append(errs, grouperror.Prefix(fmt.Sprintf("field value %+q: ", f.name), err))
			continue
		}
		err = setter.Set(&result, f.name, fieldVal, convertArgs)
		if err != nil {
			errs = append(errs, grouperror.Prefix(fmt.Sprintf("set field %+q: ", f.name), err))
		}
	}
	return result, grouperror.Join(errs...)
}

func (c *Container) executeServiceCalls(
	result any,
	svc Service,
	contextualBag keyValue,
) (any, error) {
	var errs []error

	for _, call := range svc.calls {
		action := "call"
		if call.wither {
			action = "wither"
		}

		params, err := c.resolveDeps(contextualBag, call.deps...)
		if err != nil {
			errs = append(errs, grouperror.Prefix(fmt.Sprintf("resolve args %+q: ", call.method), err))
			continue
		}

		if call.wither {
			result, err = caller.ForceCallWitherByName(&result, call.method, params, convertArgs)
			if err != nil {
				errs = append(errs, grouperror.Prefix(fmt.Sprintf("%s %+q: ", action, call.method), err))
				// wither may return a nil value for error,
				// so we have to stop execution here
				break
			}
		} else {
			_, err = caller.ForceCallByName(&result, call.method, params, convertArgs)
			if err != nil {
				errs = append(errs, grouperror.Prefix(fmt.Sprintf("%s %+q: ", action, call.method), err))
			}
		}
	}

	return result, grouperror.Join(errs...)
}

func (c *Container) decorateService(
	id string,
	result any,
	svc Service,
	contextualBag keyValue,
) (any, error) {
	// for decorators, we have to stop execution on the very first error,
	// because in case of error they may return a nil-value
	for i, dec := range c.decorators {
		if _, tagged := svc.tags[dec.tag]; !tagged {
			continue
		}
		payload := DecoratorPayload{
			Tag:       dec.tag,
			ServiceID: id,
			Service:   result,
		}
		params, err := c.resolveDeps(contextualBag, dec.deps...)
		if err != nil {
			return nil, grouperror.Prefix(fmt.Sprintf("resolve decorator args #%d: ", i), err)
		}
		params = append([]any{payload}, params...)
		result, err = caller.CallProvider(dec.fn, params, convertArgs)
		if err != nil {
			return nil, grouperror.Prefix(fmt.Sprintf("decorator #%d: ", i), err)
		}
	}

	return result, nil
}

func (c *Container) getTaggedBy(tag string, contextualBag keyValue) (result []any, err error) {
	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("getTaggedBy(%+q): ", tag), err)
		}
	}()

	services := make([]struct {
		id       string
		priority int
	}, 0)
	for id, s := range c.services {
		priority, ok := s.tags[tag]
		if !ok {
			continue
		}
		services = append(services, struct {
			id       string
			priority int
		}{
			id:       id,
			priority: priority,
		})
	}

	sort.SliceStable(services, func(i, j int) bool {
		if services[i].priority == services[j].priority {
			return services[i].id < services[j].id
		}
		return services[i].priority > services[j].priority
	})

	result = make([]any, len(services))
	var errs []error
	for i, s := range services {
		var cErr error
		result[i], cErr = c.get(s.id, contextualBag)
		if cErr != nil {
			errs = append(errs, cErr)
		}
	}

	return result, grouperror.Join(errs...)
}
