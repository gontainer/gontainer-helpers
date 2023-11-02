// Copyright (c) 2023 Bartłomiej Krukowski
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

package graph

import (
	pkgGraph "github.com/gontainer/gontainer-helpers/v3/graph"
)

type graph interface {
	Deps(n string) []string
	AddDep(from, to string)
	CircularDeps() [][]string
}

type dependencyGraph struct {
	graph        graph
	dependencies dependencies
}

func New() *dependencyGraph {
	return &dependencyGraph{
		graph:        pkgGraph.New(),
		dependencies: make(dependencies),
	}
}

func (d *dependencyGraph) AddService(serviceID string, tags []string) {
	svc := d.dependencies.service(serviceID)
	for _, t := range tags {
		d.graph.AddDep(d.dependencies.tag(t).id, svc.id)
		d.graph.AddDep(svc.id, d.dependencies.decoratedByTag(t).id)
	}
}

func (d *dependencyGraph) ServiceDependsOnServices(serviceID string, dependenciesIDs []string) {
	svc := d.dependencies.service(serviceID)
	for _, dID := range dependenciesIDs {
		d.graph.AddDep(svc.id, d.dependencies.service(dID).id)
	}
}

func (d *dependencyGraph) ServiceDependsOnParams(serviceID string, dependenciesIDs []string) {
	svc := d.dependencies.service(serviceID)
	for _, dID := range dependenciesIDs {
		d.graph.AddDep(svc.id, d.dependencies.param(dID).id)
	}
}

func (d *dependencyGraph) ServiceDependsOnTags(serviceID string, tagsIDs []string) {
	svc := d.dependencies.service(serviceID)
	for _, tID := range tagsIDs {
		d.graph.AddDep(svc.id, d.dependencies.tag(tID).id)
	}
}

func (d *dependencyGraph) AddDecorator(decoratorID int, tag string) {
	d.graph.AddDep(
		d.dependencies.decoratedByTag(tag).id,
		d.dependencies.decorator(decoratorID).id,
	)
}

func (d *dependencyGraph) DecoratorDependsOnServices(decoratorID int, dependenciesIDs []string) {
	dec := d.dependencies.decorator(decoratorID)
	for _, dID := range dependenciesIDs {
		d.graph.AddDep(dec.id, d.dependencies.service(dID).id)
	}
}

func (d *dependencyGraph) DecoratorDependsOnParams(decoratorID int, dependenciesIDs []string) {
	dec := d.dependencies.decorator(decoratorID)
	for _, dID := range dependenciesIDs {
		d.graph.AddDep(dec.id, d.dependencies.param(dID).id)
	}
}

func (d *dependencyGraph) DecoratorDependsOnTags(decoratorID int, tagsIDs []string) {
	dec := d.dependencies.decorator(decoratorID)
	for _, tID := range tagsIDs {
		d.graph.AddDep(dec.id, d.dependencies.tag(tID).id)
	}
}

func (d *dependencyGraph) ParamDependsOnParam(paramID string, dependencyID string) {
	d.graph.AddDep(
		d.dependencies.param(paramID).id,
		d.dependencies.param(dependencyID).id,
	)
}

// Deps returns a list of all (direct and indirect) dependencies for the given service
func (d *dependencyGraph) Deps(serviceID string) []Dependency {
	deps := d.graph.Deps(d.dependencies.service(serviceID).id)
	r := make([]Dependency, len(deps))
	for i, cd := range deps {
		r[i] = d.dependencies[cd]
	}
	return r
}

func (d *dependencyGraph) CircularDeps() [][]Dependency {
	graphCircularDeps := d.graph.CircularDeps()
	circularDeps := make([][]Dependency, len(graphCircularDeps))
	for i, line := range graphCircularDeps {
		circularDeps[i] = make([]Dependency, len(line))
		for j, cd := range line {
			circularDeps[i][j] = d.dependencies[cd]
		}
	}

	for i, cycle := range circularDeps {
		circularDeps[i] = normalizeCycle(cycle)
	}

	return circularDeps
}

// normalizeCycle finds a service with the lowest name, and put it on the beginning of the cycle.
// [b, c, a, b] => [a, b, c, a]
func normalizeCycle(cycle []Dependency) []Dependency {
	lowest := 0

	for indexNode, node := range cycle { // find the first service
		if node.kind == dependencyService {
			lowest = indexNode
			break
		}
	}

	for indexNode, node := range cycle { // find the lowest service name
		if node.kind != dependencyService {
			continue
		}
		if cycle[lowest].Resource > cycle[indexNode].Resource {
			lowest = indexNode
		}
	}

	for i := 0; i < lowest; i++ {
		cycle = append(cycle[1:], cycle[1])
	}

	return cycle
}
