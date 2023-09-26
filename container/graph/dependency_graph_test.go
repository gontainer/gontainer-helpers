package graph_test

import (
	"testing"

	"github.com/gontainer/gontainer-helpers/container/graph"
	errAssert "github.com/gontainer/gontainer-helpers/errors/assert"
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

		expected := []string{
			`@company -> @department -> @team -> @hr -> !tagged organization -> @holding -> @company`,
			`@company -> @department -> @team -> @hr -> !tagged organization -> @company`,
			`@holding -> @holding`,
			`@department -> @team -> @department`,
			`@chro -> @hr -> @chro`,
			`@hr -> @hr`,
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
