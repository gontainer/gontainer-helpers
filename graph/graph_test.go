// Copyright (c) 2023-2024 Bartłomiej Krukowski
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
