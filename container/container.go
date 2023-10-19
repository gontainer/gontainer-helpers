package container

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/gontainer/gontainer-helpers/caller"
	"github.com/gontainer/gontainer-helpers/container/internal/groupcontext"
	"github.com/gontainer/gontainer-helpers/grouperror"
)

type serviceDecorator struct {
	tag  string
	fn   any
	deps []Dependency
}

type container struct {
	graphBuilder   *graphBuilder
	services       map[string]Service
	cacheShared    keyValue
	serviceLockers map[string]sync.Locker
	globalLocker   rwlocker
	decorators     []serviceDecorator
	groupContext   interface {
		Add(context.Context)
		Wait()
	}
	id ctxKey
}

type ctxKey uint64

var currentContainerID = new(uint64)

// NewContainer creates a concurrent-safe DI container.
func NewContainer() *container {
	c := &container{
		services:       make(map[string]Service),
		cacheShared:    newSafeMap(),
		serviceLockers: make(map[string]sync.Locker),
		globalLocker:   &sync.RWMutex{},
		groupContext:   groupcontext.New(),
		id:             ctxKey(atomic.AddUint64(currentContainerID, 1)),
	}
	c.graphBuilder = newGraphBuilder(c)
	return c
}

func (c *container) CircularDeps() error {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	return grouperror.Prefix("container.CircularDeps(): ", c.graphBuilder.circularDeps())
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
		if _, ok := c.serviceLockers[id]; !ok {
			c.serviceLockers[id] = &sync.Mutex{}
		}
	default:
		delete(c.serviceLockers, id)
	}
}

func (c *container) AddDecorator(tag string, decorator any, deps ...Dependency) {
	c.globalLocker.Lock()
	defer c.globalLocker.Unlock()

	c.graphBuilder.invalidate()

	c.decorators = append(c.decorators, serviceDecorator{
		tag:  tag,
		fn:   decorator,
		deps: deps,
	})
}

func (c *container) contextBag(ctx context.Context) keyValue {
	bag := ctx.Value(c.id)
	if bag == nil {
		panic("the given context is not attached to the given container, call `ctx = container.ContextWithContainer(ctx, c)`")
	}
	return bag.(keyValue)
}

func (c *container) GetInContext(ctx context.Context, id string) (any, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	return c.get(id, c.contextBag(ctx))
}

func (c *container) Get(id string) (any, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	return c.get(id, newSafeMap())
}

func (c *container) resolveDeps(contextualBag keyValue, deps ...Dependency) ([]any, error) {
	r := make([]any, len(deps))
	errs := make([]error, len(deps))

	for i, d := range deps {
		var err error
		r[i], err = c.resolveDep(contextualBag, d)
		errs[i] = grouperror.Prefix(fmt.Sprintf("arg #%d: ", i), err)
	}

	return r, grouperror.Join(errs...)
}

func (c *container) resolveDep(contextualBag keyValue, d Dependency) (any, error) {
	switch d.type_ {
	case dependencyNil:
		return d.value, nil
	case dependencyTag:
		return c.getTaggedBy(d.tagID, contextualBag)
	case dependencyService:
		return c.get(d.serviceID, contextualBag)
	case dependencyProvider:
		return caller.CallProvider(d.provider, nil, convertParams)
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

func (c *container) GetTaggedBy(tag string) ([]any, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	return c.getTaggedBy(tag, newSafeMap())
}

func (c *container) GetTaggedByInContext(ctx context.Context, tag string) ([]any, error) {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	return c.getTaggedBy(tag, c.contextBag(ctx))
}

func (c *container) getTaggedBy(tag string, contextualBag keyValue) (result []any, err error) {
	defer func() {
		if err != nil {
			err = grouperror.Prefix(fmt.Sprintf("container.getTaggedBy(%+q): ", tag), err)
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
