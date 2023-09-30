package container

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/caller"
	"github.com/gontainer/gontainer-helpers/errors"
	"github.com/gontainer/gontainer-helpers/setter"
)

func (c *container) get(id string, contextualBag map[string]interface{}) (result interface{}, err error) {
	defer func() {
		if err != nil {
			err = errors.PrefixedGroup(fmt.Sprintf("container.get(%+q): ", id), err)
		}
	}()

	err = c.graphBuilder.serviceCircularDeps(id)
	if err != nil {
		return nil, errors.PrefixedGroup("circular dependencies: ", err)
	}

	svc, ok := c.services[id]
	if !ok {
		return nil, errors.New("service does not exist")
	}

	currentScope := svc.scope
	if currentScope == scopeDefault {
		currentScope = c.graphBuilder.resolveScope(id)
	}
	if currentScope == scopeShared { // write operation only for scopeShared
		c.lockers[id].Lock()
		defer c.lockers[id].Unlock()
	}

	// scopeShared
	if s, cached := c.cacheShared[id]; cached {
		return s, nil
	}

	// scopeContextual: check whether the s is already created, if yes, return it
	if s, cached := contextualBag[id]; cached {
		return s, nil
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

	switch currentScope {
	case scopeContextual:
		// cache the given object only in the given context,
		// the cache will be destroyed after returning the root Service
		contextualBag[id] = result
	case scopeShared:
		// the given instance is cached, and it will be re-used each time you call `container.Call(id)`
		c.cacheShared[id] = result
	}

	return result, nil
}

func (c *container) createNewService(svc Service, contextualBag map[string]interface{}) (interface{}, error) {
	result := svc.value

	if svc.constructor != nil {
		params, err := c.resolveDeps(contextualBag, svc.constructorDeps...)
		if err != nil {
			return nil, errors.PrefixedGroup("constructor args: ", err)
		}
		result, err = caller.CallProvider(svc.constructor, params...)
		if err != nil {
			return nil, errors.PrefixedGroup("constructor: ", err)
		}
	}

	return result, nil
}

func (c *container) setServiceFields(
	result interface{},
	svc Service,
	contextualBag map[string]interface{},
) (interface{}, error) {
	errs := make([]error, len(svc.fields))
	for i, f := range svc.fields {
		fieldVal, err := c.resolveDep(contextualBag, f.dep)
		if err != nil {
			errs[i] = errors.PrefixedGroup(fmt.Sprintf("field value %+q: ", f.name), err)
			continue
		}
		err = setter.Set(&result, f.name, fieldVal)
		if err != nil {
			errs[i] = errors.PrefixedGroup(fmt.Sprintf("set field %+q: ", f.name), err)
		}
	}
	return result, errors.Group(errs...)
}

func (c *container) executeServiceCalls(
	result interface{},
	svc Service,
	contextualBag map[string]interface{},
) (interface{}, error) {
	for _, call := range svc.calls {
		action := "call"
		if call.wither {
			action = "wither"
		}

		params, err := c.resolveDeps(contextualBag, call.deps...)
		if err != nil {
			return nil, errors.PrefixedGroup(fmt.Sprintf("resolve args %+q: ", call.method), err)
		}

		if call.wither {
			result, err = caller.CallWitherByName(result, call.method, params...)
			if err != nil {
				return nil, errors.PrefixedGroup(fmt.Sprintf("%s %+q: ", action, call.method), err)
			}
		} else {
			_, err = caller.CallByName(&result, call.method, params...)
			if err != nil {
				return nil, errors.PrefixedGroup(fmt.Sprintf("%s %+q: ", action, call.method), err)
			}
		}
	}

	return result, nil
}

func (c *container) decorateService(
	id string,
	result interface{},
	svc Service,
	contextualBag map[string]interface{},
) (interface{}, error) {
	for i, dec := range c.decorators {
		if _, tagged := svc.tags[dec.tag]; !tagged {
			continue
		}
		ctx := DecoratorContext{
			Tag:       dec.tag,
			ServiceID: id,
			Service:   result,
		}
		params, err := c.resolveDeps(contextualBag, dec.deps...)
		if err != nil {
			return nil, errors.PrefixedGroup(fmt.Sprintf("resolve decorator args #%d: ", i), err)
		}
		params = append([]interface{}{ctx}, params...)
		result, err = caller.CallProvider(dec.fn, params...)
		if err != nil {
			return nil, errors.PrefixedGroup(fmt.Sprintf("decorator #%d: ", i), err)
		}
	}

	return result, nil
}
