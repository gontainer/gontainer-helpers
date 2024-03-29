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
	"fmt"

	containerGraph "github.com/gontainer/gontainer-helpers/v3/container/internal/graph"
	"github.com/gontainer/gontainer-helpers/v3/container/internal/maps"
)

// graphBuilder is a helper for [*Container], it analyzes dependencies to resolve the scope in runtime,
// and detect circular dependencies.
// It is not concurrent-safe.
type graphBuilder struct {
	container            *Container
	servicesCycles       map[string][]int
	paramsCycles         map[string][]int
	scopes               map[string]scope
	computedCircularDeps [][]containerGraph.Dependency
}

func newGraphBuilder(c *Container) *graphBuilder {
	return &graphBuilder{
		container: c,
	}
}

// invalidate invalidates the cache generated by the method warmUp.
func (g *graphBuilder) invalidate() {
	g.servicesCycles = nil
	g.paramsCycles = nil
	g.scopes = nil
	g.computedCircularDeps = nil
}

func (g *graphBuilder) warmUpCircularDeps() {
	g.servicesCycles = make(map[string][]int)
	g.paramsCycles = make(map[string][]int)

	for cycleID, cycle := range g.computedCircularDeps {
		// first and last elements in the cycle points to the same dependency, so we should ignore one of them
		// a -> b -> c -> a
		for _, dep := range cycle[1:] {
			if dep.IsService() {
				g.servicesCycles[dep.Resource] = append(g.servicesCycles[dep.Resource], cycleID)
				continue
			}
			if dep.IsParam() {
				g.paramsCycles[dep.Resource] = append(g.paramsCycles[dep.Resource], cycleID)
				continue
			}
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
	graph := containerGraph.New()

	// iterate over `g.Container.services` in the same order always,
	// otherwise we would add elements to the tree in different order
	// it may lead to having inconsistent results in the method `CircularDeps()`
	for _, sID := range maps.SortedStringKeys(g.container.services) {
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

		dependenciesServices, dependenciesParams, dependenciesTags := depsToRawServicesParamsTags(deps...)
		graph.ServiceDependsOnServices(sID, dependenciesServices)
		graph.ServiceDependsOnParams(sID, dependenciesParams)
		graph.ServiceDependsOnTags(sID, dependenciesTags)
	}

	for dID, d := range g.container.decorators {
		graph.AddDecorator(dID, d.tag)

		dependenciesServices, dependenciesParams, dependenciesTags := depsToRawServicesParamsTags(d.deps...)
		graph.DecoratorDependsOnServices(dID, dependenciesServices)
		graph.DecoratorDependsOnParams(dID, dependenciesParams)
		graph.DecoratorDependsOnTags(dID, dependenciesTags)
	}

	for _, pID := range maps.SortedStringKeys(g.container.params) {
		dep := g.container.params[pID]
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
	s, ok := g.scopes[serviceID]
	if !ok {
		panic(fmt.Sprintf("scope for %+q does not exist in cache", serviceID))
	}

	return s
}

func (g *graphBuilder) circularDeps() error {
	return containerGraph.CircularDepsToError(g.computedCircularDeps)
}

func (g *graphBuilder) serviceCircularDeps(serviceID string) error {
	circularDeps := make([][]containerGraph.Dependency, 0, len(g.servicesCycles[serviceID]))
	for _, cycleID := range g.servicesCycles[serviceID] {
		circularDeps = append(circularDeps, g.computedCircularDeps[cycleID])
	}

	return containerGraph.CircularDepsToError(circularDeps)
}

func (g *graphBuilder) paramCircularDeps(paramID string) error {
	circularDeps := make([][]containerGraph.Dependency, 0, len(g.paramsCycles[paramID]))
	for _, cycleID := range g.paramsCycles[paramID] {
		circularDeps = append(circularDeps, g.computedCircularDeps[cycleID])
	}

	return containerGraph.CircularDepsToError(circularDeps)
}

func depsToRawServicesParamsTags(deps ...Dependency) (services, params, tags []string) {
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
