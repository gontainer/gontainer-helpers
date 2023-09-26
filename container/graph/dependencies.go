package graph

import (
	"fmt"
)

type dependencies map[string]Dependency

// creates and returns a service Dependency. Returns the existing Dependency if exists.
//
//	@service
func (d dependencies) service(n string) Dependency {
	dep := Dependency{
		id:       fmt.Sprintf("service(%s)", n),
		Resource: n,
		kind:     dependencyService,
		Pretty:   fmt.Sprintf("@%s", n),
	}
	d[dep.id] = dep
	return dep
}

// creates and returns a tag Dependency. Returns the existing Dependency if exists.
//
//	tag(http.handler)
func (d dependencies) tag(n string) Dependency {
	dep := Dependency{
		id:       fmt.Sprintf("tag(%s)", n),
		Resource: n,
		kind:     dependencyTag,
		Pretty:   fmt.Sprintf("!tagged %s", n),
	}
	d[dep.id] = dep
	return dep
}

// creates and returns a decorated by tag Dependency. Returns the existing Dependency if exists.
//
//	decorate(!tagged http.handler)
func (d dependencies) decoratedByTag(n string) Dependency {
	dep := Dependency{
		id:       fmt.Sprintf("decorate(%s)", n),
		Resource: n,
		kind:     dependencyDecoratedByTag,
		Pretty:   fmt.Sprintf("decorate(!tagged %s)", n),
	}
	d[dep.id] = dep
	return dep
}

// creates and returns a decorator Dependency. Returns the existing Dependency if exists.
//
//	decorator(#0)
func (d dependencies) decorator(id int) Dependency {
	dep := Dependency{
		id:       fmt.Sprintf("decorator(#%d)", id),
		Resource: fmt.Sprintf("%d", id),
		kind:     dependencyDecorator,
		Pretty:   fmt.Sprintf("decorator(#%d)", id),
	}
	d[dep.id] = dep
	return dep
}
