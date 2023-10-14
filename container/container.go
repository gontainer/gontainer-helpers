package container

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"

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
	graphBuilder   *graphBuilder
	services       map[string]Service
	cacheShared    keyValue
	serviceLockers map[string]sync.Locker
	globalLocker   rwlocker
	decorators     []serviceDecorator
	id             ctxKey
}

type ctxKey uint64

var currentContainerID uint64

// NewContainer creates a concurrent-safe DI container.
func NewContainer() *container {
	c := &container{
		services:       make(map[string]Service),
		cacheShared:    newSafeMap(),
		serviceLockers: make(map[string]sync.Locker),
		globalLocker:   &sync.RWMutex{},
		id:             ctxKey(atomic.AddUint64(&currentContainerID, 1)),
	}
	c.graphBuilder = newGraphBuilder(c)
	return c
}

func (c *container) CircularDeps() error {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	return errors.PrefixedGroup("container.CircularDeps(): ", c.graphBuilder.circularDeps())
}

func (c *container) OverrideService(id string, s Service) {
	c.globalLocker.Lock()
	defer c.globalLocker.Unlock()

	c.graphBuilder.invalidate()

	c.services[id] = s
	c.cacheShared.delete(id)
	switch s.scope {
	case
		scopeDefault,
		scopeShared,
		scopeContextual:
		c.serviceLockers[id] = &sync.Mutex{}
	}
}

func (c *container) AddDecorator(tag string, decorator interface{}, deps ...Dependency) {
	c.globalLocker.Lock()
	defer c.globalLocker.Unlock()

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
//
// Deprecated: use copier.Copy or copier.ConvertAndCopy.
// It's been deprecated to avoid adding a complex method `CopyServiceToContext`.
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

func (c *container) contextBag(ctx context.Context) keyValue {
	bag := ctx.Value(c.id)
	if bag == nil {
		panic("the given context is not attached to the given container, call `ctx = container.ContextWithContainer(ctx, c)`")
	}
	return bag.(keyValue)
}

func (c *container) GetWithContext(ctx context.Context, id string) (interface{}, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	return c.get(id, c.contextBag(ctx))
}

func (c *container) Get(id string) (interface{}, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	return c.get(id, newSafeMap())
}

func (c *container) resolveDeps(contextualBag keyValue, deps ...Dependency) ([]interface{}, error) {
	r := make([]interface{}, len(deps))
	errs := make([]error, len(deps))

	for i, d := range deps {
		var err error
		r[i], err = c.resolveDep(contextualBag, d)
		errs[i] = errors.PrefixedGroup(fmt.Sprintf("arg #%d: ", i), err)
	}

	return r, errors.Group(errs...)
}

func (c *container) resolveDep(contextualBag keyValue, d Dependency) (interface{}, error) {
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
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	s, exists := c.services[id]
	if !exists {
		return false
	}
	_, ok := s.tags[tag]
	return ok
}

func (c *container) GetTaggedBy(tag string) ([]interface{}, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	return c.getTaggedBy(tag, newSafeMap())
}

func (c *container) GetTaggedByWithContext(ctx context.Context, tag string) ([]interface{}, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	return c.getTaggedBy(tag, c.contextBag(ctx))
}

func (c *container) getTaggedBy(tag string, contextualBag keyValue) (result []interface{}, err error) {
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
			return services[i].id < services[j].id
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
