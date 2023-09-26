package container

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/gontainer/gontainer-helpers/caller"
	"github.com/gontainer/gontainer-helpers/copier"
	"github.com/gontainer/gontainer-helpers/errors"
	"github.com/gontainer/gontainer-helpers/setter"
)

type serviceDecorator struct {
	tag  string
	fn   interface{}
	deps []Dependency
}

type container struct {
	graphBuilder *graphBuilder
	services     map[string]Service
	cacheShared  map[string]interface{}
	lockers      map[string]sync.Locker
	rwlocker     rwlocker
	decorators   []serviceDecorator
}

func NewContainer() *container {
	c := &container{
		services:    make(map[string]Service),
		cacheShared: make(map[string]interface{}),
		lockers:     make(map[string]sync.Locker),
		rwlocker:    &sync.RWMutex{},
	}
	c.graphBuilder = newGraphBuilder(c)
	return c
}

func (c *container) CircularDeps() error {
	c.rwlocker.RLock()
	defer c.rwlocker.RUnlock()

	return errors.PrefixedGroup("container.CircularDeps(): ", c.graphBuilder.circularDeps())
}

func (c *container) OverrideService(id string, s Service) {
	c.rwlocker.Lock()
	defer c.rwlocker.Unlock()

	c.graphBuilder.invalidate()

	c.services[id] = s
	delete(c.cacheShared, id)
	switch s.scope {
	case
		scopeDefault,
		scopeShared:
		c.lockers[id] = &sync.Mutex{}
	}
}

func (c *container) AddDecorator(tag string, decorator interface{}, deps ...Dependency) {
	c.rwlocker.Lock()
	defer c.rwlocker.Unlock()

	c.graphBuilder.invalidate()

	c.decorators = append(c.decorators, serviceDecorator{
		tag:  tag,
		fn:   decorator,
		deps: deps,
	})
}

// CopyServiceTo gets or creates the desired service and copies it to the given pointer
//
//	var server *http.Server
//	container.CopyServiceTo("server", &server)
func (c *container) CopyServiceTo(id string, dst interface{}) (err error) {
	defer func() {
		if err != nil {
			err = errors.PrefixedGroup(fmt.Sprintf("container.CopyServiceTo(%+q): ", id), err)
		}
	}()
	r, err := c.Get(id)
	if err != nil {
		return err
	}
	return copier.Copy(r, dst)
}

func (c *container) Get(id string) (result interface{}, err error) {
	c.rwlocker.RLock()
	defer c.rwlocker.RUnlock()

	return c.get(id, make(map[string]interface{}))
}

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

	result = svc.value

	// constructor
	if svc.constructor != nil {
		var params []interface{}
		params, err = c.resolveDeps(contextualBag, svc.constructorDeps...)
		if err != nil {
			return nil, errors.PrefixedGroup("constructor args: ", err)
		}
		result, err = caller.CallProvider(svc.constructor, params...)
		if err != nil {
			return nil, errors.PrefixedGroup("constructor: ", err)
		}
	}

	// fields
	for _, f := range svc.fields {
		var fieldVal interface{}
		fieldVal, err = c.resolveDep(contextualBag, f.dep)
		if err != nil {
			return nil, errors.PrefixedGroup(fmt.Sprintf("field value %+q: ", f.name), err)
		}
		err = setter.Set(&result, f.name, fieldVal)
		if err != nil {
			return nil, errors.PrefixedGroup(fmt.Sprintf("set field %+q: ", f.name), err)
		}
	}

	// calls
	for _, call := range svc.calls {
		action := "call"
		if call.wither {
			action = "wither"
		}

		var params []interface{}
		params, err = c.resolveDeps(contextualBag, call.deps...)
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

	// decorators
	for i, dec := range c.decorators {
		if _, tagged := svc.tags[dec.tag]; !tagged {
			continue
		}
		ctx := DecoratorContext{
			Tag:       dec.tag,
			ServiceID: id,
			Service:   result,
		}
		var params []interface{}
		params, err = c.resolveDeps(contextualBag, dec.deps...)
		if err != nil {
			return nil, errors.PrefixedGroup(fmt.Sprintf("resolve decorator args #%d: ", i), err)
		}
		params = append([]interface{}{ctx}, params...)
		result, err = caller.CallProvider(dec.fn, params...)
		if err != nil {
			return nil, errors.PrefixedGroup(fmt.Sprintf("decorator #%d: ", i), err)
		}
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

func (c *container) resolveDeps(contextualBag map[string]interface{}, deps ...Dependency) ([]interface{}, error) {
	r := make([]interface{}, len(deps))
	var err error

	for i, d := range deps {
		r[i], err = c.resolveDep(contextualBag, d)
		if err != nil {
			return nil, errors.PrefixedGroup(fmt.Sprintf("arg #%d: ", i), err)
		}
	}

	return r, err
}

func (c *container) resolveDep(contextualBag map[string]interface{}, d Dependency) (interface{}, error) {
	switch d.type_ {
	case dependencyNil:
		return d.value, nil
	case dependencyTag:
		return c.getTaggedBy(d.tagID, contextualBag)
	case dependencyService:
		return c.get(d.serviceID, contextualBag)
	case dependencyProvider:
		return caller.CallProvider(d.provider)
	}

	return nil, errors.New("unknown dependency type")
}

func (c *container) IsTaggedBy(id string, tag string) bool {
	c.rwlocker.RLock()
	defer c.rwlocker.RUnlock()

	s, exists := c.services[id]
	if !exists {
		return false
	}
	_, ok := s.tags[tag]
	return ok
}

func (c *container) GetTaggedBy(tag string) (result []interface{}, err error) {
	c.rwlocker.RLock()
	defer c.rwlocker.RUnlock()

	return c.getTaggedBy(tag, make(map[string]interface{}))
}

func (c *container) getTaggedBy(tag string, contextualBag map[string]interface{}) (result []interface{}, err error) {
	defer func() {
		if err != nil {
			err = errors.PrefixedGroup(fmt.Sprintf("container.getTaggedBy(%+q): ", tag), err)
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
			return strings.Compare(services[i].id, services[j].id) < 0
		}
		return services[i].priority > services[j].priority
	})

	result = make([]interface{}, len(services))
	for i, s := range services {
		result[i], err = c.get(s.id, contextualBag)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
