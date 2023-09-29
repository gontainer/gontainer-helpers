package graph

import (
	pkgGraph "github.com/gontainer/gontainer-helpers/graph"
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

func (d *dependencyGraph) DecoratorDependsOnTags(decoratorID int, tagsIDs []string) {
	dec := d.dependencies.decorator(decoratorID)
	for _, tID := range tagsIDs {
		d.graph.AddDep(dec.id, d.dependencies.tag(tID).id)
	}
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
