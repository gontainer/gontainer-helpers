package container

import (
	"fmt"
	"sort"
	"sync"

	containerGraph "github.com/gontainer/gontainer-helpers/container/graph"
)

type graphBuilder struct {
	container            *container
	valid                bool
	locker               rwlocker
	servicesCycles       map[string][]int
	scopes               map[string]scope
	computedCircularDeps [][]containerGraph.Dependency
}

func newGraphBuilder(c *container) *graphBuilder {
	return &graphBuilder{
		container: c,
		valid:     false,
		locker:    &sync.RWMutex{},
	}
}

// invalidate invalidates the cache generated by the method warmUp.
func (g *graphBuilder) invalidate() {
	g.locker.Lock()
	defer g.locker.Unlock()

	g.valid = false
	g.servicesCycles = nil
	g.scopes = nil
	g.computedCircularDeps = nil
}

func (g *graphBuilder) warmUpCircularDeps() {
	g.servicesCycles = make(map[string][]int)
	for cycleID, cycle := range g.computedCircularDeps {
		// first and last elements in the cycle points to the same dependency, so we should ignore one of them
		// a -> b -> c -> a
		for _, dep := range cycle[1:] {
			if !dep.IsService() {
				continue
			}

			g.servicesCycles[dep.Resource] = append(g.servicesCycles[dep.Resource], cycleID)
		}
	}
}

func (g *graphBuilder) warmUpScopes(
	graph interface {
		Deps(serviceID string) []containerGraph.Dependency
	},
) {
	g.scopes = make(map[string]scope)
	for sID, s := range g.container.services {
		if s.scope != scopeDefault {
			continue
		}
		hasContextual := false
		for _, d := range graph.Deps(sID) {
			if !d.IsService() {
				continue
			}
			hasContextual = g.container.services[d.Resource].scope == scopeContextual
			if hasContextual {
				break
			}
		}
		if hasContextual {
			g.scopes[sID] = scopeContextual
		} else {
			g.scopes[sID] = scopeShared
		}
	}
}

// warmUp prepares and caches the circular deps.
func (g *graphBuilder) warmUp() {
	g.locker.Lock()
	defer g.locker.Unlock()

	if g.valid {
		return
	}

	defer func() {
		g.valid = true
	}()

	graph := containerGraph.New()

	// iterate over `g.container.services` in the same order always,
	// otherwise we would add elements to the tree in different order
	// it may lead to having inconsistent results in the method `CircularDeps()`
	sIDs := make([]string, 0, len(g.container.services))
	for sID := range g.container.services {
		sIDs = append(sIDs, sID)
	}
	sort.Strings(sIDs)

	for _, sID := range sIDs {
		s := g.container.services[sID]

		var tags []string
		for tag := range s.tags {
			tags = append(tags, tag)
		}
		graph.AddService(sID, tags)

		var deps []Dependency
		deps = append(deps, s.constructorDeps...)
		for _, call := range s.calls {
			deps = append(deps, call.deps...)
		}
		for _, field := range s.fields {
			deps = append(deps, field.dep)
		}

		dependenciesServices, dependenciesParams, dependenciesTags := depsToRawServicesTags(deps...)
		graph.ServiceDependsOnServices(sID, dependenciesServices)
		graph.ServiceDependsOnParams(sID, dependenciesParams)
		graph.ServiceDependsOnTags(sID, dependenciesTags)
	}

	for dID, d := range g.container.decorators {
		graph.AddDecorator(dID, d.tag)

		dependenciesServices, dependenciesParams, dependenciesTags := depsToRawServicesTags(d.deps...)
		graph.DecoratorDependsOnServices(dID, dependenciesServices)
		graph.DecoratorDependsOnParams(dID, dependenciesParams)
		graph.DecoratorDependsOnTags(dID, dependenciesTags)
	}

	pIDs := make([]string, 0, len(g.container.paramContainer.params))
	for pID := range g.container.paramContainer.params {
		pIDs = append(pIDs, pID)
	}
	sort.Strings(pIDs)

	for _, pID := range pIDs {
		dep := g.container.paramContainer.params[pID]
		if dep.type_ == dependencyParam {
			graph.ParamDependsOnParam(pID, dep.paramID)
		}
	}

	g.computedCircularDeps = graph.CircularDeps()
	g.warmUpCircularDeps()
	g.warmUpScopes(graph)
}

// resolveScope returns scopeContextual when at least on dependency is contextual,
// otherwise it returns scopeShared. It works for service with the scope scopeDefault.
func (g *graphBuilder) resolveScope(serviceID string) scope {
	g.warmUp()

	g.locker.RLock()
	defer g.locker.RUnlock()

	s, ok := g.scopes[serviceID]
	if !ok {
		panic(fmt.Sprintf("scope for %+q does not exist in cache", serviceID))
	}

	return s
}

func (g *graphBuilder) circularDeps() error {
	g.warmUp()

	g.locker.RLock()
	defer g.locker.RUnlock()

	return containerGraph.CircularDepsToError(g.computedCircularDeps)
}

func (g *graphBuilder) serviceCircularDeps(serviceID string) error {
	g.warmUp()

	g.locker.RLock()
	defer g.locker.RUnlock()

	var circularDeps [][]containerGraph.Dependency
	for _, cycleID := range g.servicesCycles[serviceID] {
		circularDeps = append(circularDeps, g.computedCircularDeps[cycleID])
	}

	return containerGraph.CircularDepsToError(circularDeps)
}

func depsToRawServicesTags(deps ...Dependency) (services, params, tags []string) {
	for _, dep := range deps {
		switch dep.type_ {
		case dependencyService:
			services = append(services, dep.serviceID)
		case dependencyParam:
			params = append(params, dep.paramID)
		case dependencyTag:
			tags = append(tags, dep.tagID)
		}
	}
	return
}
