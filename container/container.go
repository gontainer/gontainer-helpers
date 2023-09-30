package container

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/gontainer/gontainer-helpers/caller"
	"github.com/gontainer/gontainer-helpers/copier"
	"github.com/gontainer/gontainer-helpers/errors"
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

func (c *container) Get(id string) (interface{}, error) {
	c.rwlocker.RLock()
	defer c.rwlocker.RUnlock()

	return c.get(id, make(map[string]interface{}))
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

func (c *container) GetTaggedBy(tag string) ([]interface{}, error) {
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
