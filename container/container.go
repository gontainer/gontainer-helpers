package container

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/gontainer/gontainer-helpers/v2/caller"
	"github.com/gontainer/gontainer-helpers/v2/container/internal/groupcontext"
	"github.com/gontainer/gontainer-helpers/v2/grouperror"
)

// Container is a DI container. Use [New] to allocate a new instance.
type Container struct {
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
	onceWarmUp    *sync.Once
	id            ctxKey
}

type serviceDecorator struct {
	tag  string
	fn   any
	deps []Dependency
}

type ctxKey uint64

var (
	currentContainerID = new(uint64)
)

/*
New creates a concurrent-safe DI Container.

	type Person struct {
		Name string
	}

	s := container.NewService()
	s.SetConstructor(
		func(n string) Person {
			return Person{
				Name: n,
			}
		},
		container.NewDependencyParam("name"),
	)

	c := container.New()
	c.OverrideService("jane", s)
	c.OverrideParam("name", container.NewDependencyValue("Jane"))

	jane, _ := c.Get("jane")
	fmt.Printf("%+v\n", jane)
	// Output: {Name:Jane}
*/
func New() *Container {
	c := &Container{
		services:            make(map[string]Service),
		cacheSharedServices: newSafeMap(),
		serviceLockers:      make(map[string]sync.Locker),
		params:              make(map[string]Dependency),
		cacheParams:         newSafeMap(),
		paramsLockers:       make(map[string]sync.Locker),
		globalLocker:        &sync.RWMutex{},
		groupContext:        groupcontext.New(),
		contextLocker:       &sync.RWMutex{},
		onceWarmUp:          &sync.Once{},
		id:                  ctxKey(atomic.AddUint64(currentContainerID, 1)),
	}
	c.graphBuilder = newGraphBuilder(c)
	return c
}

// CircularDeps returns an error if there is any circular dependency.
func (c *Container) CircularDeps() error {
	c.globalLocker.RLock()
	defer c.globalLocker.RUnlock()

	c.warmUpGraph()

	return grouperror.Prefix("CircularDeps(): ", c.graphBuilder.circularDeps())
}

func (c *Container) resolveDeps(contextualBag keyValue, deps ...Dependency) ([]any, error) {
	r := make([]any, len(deps))
	errs := make([]error, len(deps))

	for i, d := range deps {
		var err error
		r[i], err = c.resolveDep(contextualBag, d)
		errs[i] = grouperror.Prefix(fmt.Sprintf("arg #%d: ", i), err)
	}

	return r, grouperror.Join(errs...)
}

func (c *Container) resolveDep(contextualBag keyValue, d Dependency) (any, error) {
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
		return caller.CallProvider(d.provider, nil, convertArgs)
	case dependencyContainer:
		return c, nil
	}

	return nil, errors.New("unknown dependency type")
}

func (c *Container) invalidateGraph() {
	c.onceWarmUp = &sync.Once{}
	c.graphBuilder.invalidate()
}

func (c *Container) warmUpGraph() {
	c.onceWarmUp.Do(func() {
		c.graphBuilder.warmUp()
	})
}
