// Copyright (c) 2023–present Bartłomiej Krukowski
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
	"sync"
	"sync/atomic"

	"github.com/gontainer/gontainer-helpers/v3/container/internal/groupcontext"
	"github.com/gontainer/grouperror"
	"github.com/gontainer/reflectpro/caller"
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
	onceWarmUp    interface{ Do(func()) }
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
		dependency.Param("name"),
	)

	c := container.New()
	c.OverrideService("jane", s)
	c.OverrideParam("name", dependency.Value("Jane"))

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

func (c *Container) resolveDeps(ctx context.Context, contextualBag keyValue, deps ...Dependency) ([]any, error) {
	if len(deps) == 0 {
		return nil, nil
	}
	r := make([]any, len(deps))
	var errs []error

	for i, d := range deps {
		var err error
		r[i], err = c.resolveDep(ctx, contextualBag, d)
		if err != nil {
			errs = append(errs, grouperror.Prefix(fmt.Sprintf("arg #%d: ", i), err))
		}
	}

	return r, grouperror.Join(errs...)
}

func (c *Container) resolveDep(ctx context.Context, contextualBag keyValue, d Dependency) (any, error) {
	switch d.type_ {
	case dependencyValue:
		return d.value, nil
	case dependencyTag:
		return c.getTaggedBy(ctx, d.tagID, contextualBag)
	case dependencyService:
		return c.get(ctx, d.serviceID, contextualBag)
	case dependencyParam:
		return c.getParam(d.paramID)
	case dependencyProvider:
		return caller.CallProvider(d.provider, nil, convertArgs)
	case dependencyContainer:
		return c, nil
	case dependencyContext:
		return ctx, nil
	}

	return nil, errors.New("unknown dependency type")
}

func (c *Container) invalidateGraph() {
	c.onceWarmUp = &sync.Once{}
	c.graphBuilder.invalidate()
}

func (c *Container) warmUpGraph() {
	c.onceWarmUp.Do(c.graphBuilder.warmUp)
}
