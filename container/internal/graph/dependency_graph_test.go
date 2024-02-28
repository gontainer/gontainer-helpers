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

package graph_test

import (
	"testing"

	"github.com/gontainer/gontainer-helpers/v3/container/internal/graph"
	errAssert "github.com/gontainer/grouperror/assert"
)

func TestNew(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		g := graph.New()

		g.AddService("holding", []string{"organization"})
		g.ServiceDependsOnServices("holding", []string{"company", "holding"})

		g.AddService("company", []string{"organization"})
		g.ServiceDependsOnServices("company", []string{"department"})

		g.AddService("department", []string{})
		g.ServiceDependsOnServices("department", []string{"team"})

		g.AddService("team", []string{})
		g.ServiceDependsOnServices("team", []string{"hr", "department"})

		g.AddService("hr", []string{})
		g.ServiceDependsOnServices("hr", []string{"chro", "hr"})
		g.ServiceDependsOnTags("hr", []string{"organization"})

		g.AddService("chro", []string{})
		g.ServiceDependsOnServices("chro", []string{"hr"})

		g.ParamDependsOnParam("firstname", "name")
		g.ParamDependsOnParam("name", "fullname")
		g.ParamDependsOnParam("fullname", "firstname")

		expected := []string{
			`@company -> @department -> @team -> @hr -> !tagged organization -> @holding -> @company`,
			`@company -> @department -> @team -> @hr -> !tagged organization -> @company`,
			`@holding -> @holding`,
			`@department -> @team -> @department`,
			`@chro -> @hr -> @chro`,
			`@hr -> @hr`,
			`%firstname% -> %name% -> %fullname% -> %firstname%`,
		}

		err := graph.CircularDepsToError(g.CircularDeps())
		errAssert.EqualErrorGroup(t, err, expected)
	})

	t.Run("Decorators", func(t *testing.T) {
		g := graph.New()

		g.AddService("db", []string{"sql.DB"})

		g.AddService("serviceA", []string{"tagB"})

		g.AddDecorator(0, "sql.DB")
		g.DecoratorDependsOnServices(0, []string{"db"})

		g.AddDecorator(1, "tagB")
		g.DecoratorDependsOnTags(1, []string{"tagB"})

		expected := []string{
			`@db -> decorate(!tagged sql.DB) -> decorator(#0) -> @db`,
			`@serviceA -> decorate(!tagged tagB) -> decorator(#1) -> !tagged tagB -> @serviceA`,
		}

		err := graph.CircularDepsToError(g.CircularDeps())
		errAssert.EqualErrorGroup(t, err, expected)
	})
}
