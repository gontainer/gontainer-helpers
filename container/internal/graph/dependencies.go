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

// creates and returns a param Dependency. Returns the existing Dependency if exists.
//
//	@service
func (d dependencies) param(n string) Dependency {
	dep := Dependency{
		id:       fmt.Sprintf("param(%s)", n),
		Resource: n,
		kind:     dependencyParam,
		Pretty:   "%" + n + "%",
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
