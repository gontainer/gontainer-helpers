package container

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/gontainer/gontainer-helpers/caller"
	"github.com/gontainer/gontainer-helpers/grouperror"
	"github.com/gontainer/gontainer-helpers/setter"
)

func (c *Container) Get(id string) (any, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	return c.get(id, newSafeMap())
}

func (c *Container) GetInContext(ctx context.Context, id string) (any, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	return c.get(id, c.contextBag(ctx))
}

func (c *Container) GetTaggedBy(tag string) ([]any, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	return c.getTaggedBy(tag, newSafeMap())
}

func (c *Container) GetTaggedByInContext(ctx context.Context, tag string) ([]any, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	return c.getTaggedBy(tag, c.contextBag(ctx))
}

func (c *Container) IsTaggedBy(id string, tag string) bool {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	s, exists := c.services[id]
	if !exists {
		return false
	}
	_, ok := s.tags[tag]
	return ok
}

func (c *Container) get(id string, contextualBag keyValue) (result any, err error) {
	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("Container.get(%+q): ", id), err)
		}
	}()

	err = c.graphBuilder.serviceCircularDeps(id)
	if err != nil {
		return nil, grouperror.Prefix("circular dependencies: ", err)
	}

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
	errs := make([]error, len(svc.fields))
	for i, f := range svc.fields {
		fieldVal, err := c.resolveDep(contextualBag, f.dep)
		if err != nil {
			errs[i] = grouperror.Prefix(fmt.Sprintf("field value %+q: ", f.name), err)
			continue
		}
		err = setter.Set(&result, f.name, fieldVal, convertArgs)
		if err != nil {
			errs[i] = grouperror.Prefix(fmt.Sprintf("set field %+q: ", f.name), err)
		}
	}
	return result, grouperror.Join(errs...)
}

func (c *Container) executeServiceCalls(
	result any,
	svc Service,
	contextualBag keyValue,
) (any, error) {
	errs := make([]error, len(svc.calls))

	for i, call := range svc.calls {
		action := "call"
		if call.wither {
			action = "wither"
		}

		params, err := c.resolveDeps(contextualBag, call.deps...)
		if err != nil {
			errs[i] = grouperror.Prefix(fmt.Sprintf("resolve args %+q: ", call.method), err)
			continue
		}

		if call.wither {
			result, err = caller.CallWitherByName(result, call.method, params, convertArgs)
			if err != nil {
				errs[i] = grouperror.Prefix(fmt.Sprintf("%s %+q: ", action, call.method), err)
				// wither may return a nil value for error,
				// so we have to stop execution here
				break
			}
		} else {
			_, err = caller.CallByName(&result, call.method, params, convertArgs)
			errs[i] = grouperror.Prefix(fmt.Sprintf("%s %+q: ", action, call.method), err)
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
			err = grouperror.Prefix(fmt.Sprintf("Container.getTaggedBy(%+q): ", tag), err)
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
	for i, s := range services {
		result[i], err = c.get(s.id, contextualBag)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
