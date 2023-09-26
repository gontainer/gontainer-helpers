package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gonumGraph "gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

func Test_graph_CircularDeps(t *testing.T) {
	for i := 0; i < 100; i++ {
		d := New()
		d.AddDep("holding", "company")
		d.AddDep("company", "department")
		d.AddDep("department", "holding")
		d.AddDep("holding", "holding")
		d.AddDep("department", "department")
		d.AddDep("holding", "department")
		expected := [][]string{
			{"holding", "company", "department", "holding"},
			{"holding", "department", "holding"},
			{"holding", "holding"},
			{"department", "department"},
		}
		require.Equal(t, expected, d.CircularDeps())
	}
}

func Test_sortCycle(t *testing.T) {
	cycle := func(vals ...int64) []gonumGraph.Node {
		r := make([]gonumGraph.Node, 0, len(vals))
		for _, v := range vals {
			r = append(r, simple.Node(v))
		}
		return r
	}

	for i := 0; i < 100; i++ {
		cycles := [][]gonumGraph.Node{
			cycle(0, 1, 2),
			cycle(0, 1),
			cycle(1, 2, 3, 4),
			cycle(0, 0, 0, 0),
		}
		sortCycle(cycles)
		expected := [][]gonumGraph.Node{
			cycle(0, 0, 0, 0),
			cycle(0, 1),
			cycle(0, 1, 2),
			cycle(1, 2, 3, 4),
		}
		assert.Equal(t, expected, cycles)
	}
}

func Test_graph_Deps(t *testing.T) {
	d := New()
	d.AddDep("a", "z")
	d.AddDep("a", "a")
	d.AddDep("z", "b")
	d.AddDep("b", "c")
	d.AddDep("c", "d")
	d.AddDep("d", "e")
	d.AddDep("c", "a")
	for i := 0; i < 2; i++ {
		assert.Equal(t, []string{"b", "c", "d", "e", "z"}, d.Deps("a"))
		assert.Equal(t, []string{"a", "b", "d", "e", "z"}, d.Deps("c"))
	}
}
