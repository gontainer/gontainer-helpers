package container

import (
	"context"
	"errors"
	"fmt"
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
	graphBuilder interface {
		warmUp()
		invalidate()
		circularDeps() error
		serviceCircularDeps(serviceID string) error
		paramCircularDeps(paramID string) error
		resolveScope(serviceID string) scope
	}
	services            map[string]Service
	cacheSharedServices keyValue
	serviceLockers      map[string]sync.Locker
	params              map[string]Dependency
	cacheParams         keyValue
	paramsLockers       map[string]sync.Locker
	globalLocker        rwlocker
	decorators          []serviceDecorator
	groupContext        interface {
		Add(context.Context)
		Wait()
	}
	contextLocker rwlocker
	id            ctxKey
}

type ctxKey uint64

var currentContainerID = new(uint64)

// New creates a concurrent-safe DI container.
func New() *container {
	return NewContainer()
}

// NewContainer creates a concurrent-safe DI container.
//
// TODO: remove it, use New
func NewContainer() *container {
	c := &container{
		services:            make(map[string]Service),
		cacheSharedServices: newSafeMap(),
		serviceLockers:      make(map[string]sync.Locker),
		params:              make(map[string]Dependency),
		cacheParams:         newSafeMap(),
		paramsLockers:       make(map[string]sync.Locker),
		globalLocker:        &sync.RWMutex{},
		groupContext:        groupcontext.New(),
		contextLocker:       &sync.RWMutex{},
		id:                  ctxKey(atomic.AddUint64(currentContainerID, 1)),
	}
	c.graphBuilder = newGraphBuilder(c)
	return c
}

func (c *container) CircularDeps() error {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	return grouperror.Prefix("container.CircularDeps(): ", c.graphBuilder.circularDeps())
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
	case dependencyValue:
		return d.value, nil
	case dependencyTag:
		return c.getTaggedBy(d.tagID, contextualBag)
	case dependencyService:
		return c.get(d.serviceID, contextualBag)
	case dependencyParam:
		return c.getParam(d.paramID)
	case dependencyProvider:
		return caller.CallProvider(d.provider, nil, convertParams)
	}

	return nil, errors.New("unknown dependency type")
}
