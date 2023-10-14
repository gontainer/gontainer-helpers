package graph

import (
	"sort"

	gonumGraph "gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
	"gonum.org/v1/gonum/graph/traverse"
)

type graph struct {
	namesNodes map[string]simple.Node
	nodesNames map[simple.Node]string
	graph      *simple.DirectedGraph
	current    simple.Node
}

// Deps returns all dependencies to the given node excluding itself in the lexical order.
func (f *graph) Deps(n string) []string {
	node := f.nodeByName(n)
	var deps []string
	df := traverse.DepthFirst{
		Visit: func(curr gonumGraph.Node) {
			if node.ID() == curr.ID() {
				return
			}
			name, ok := f.nameByNode(simple.Node(curr.ID()))
			if !ok {
				return
			}
			deps = append(deps, name)
		},
		Traverse: nil,
	}
	df.Walk(f.graph, node, nil)
	sort.SliceStable(deps, func(i, j int) bool {
		return deps[i] < deps[j]
	})
	return deps
}

func (f *graph) AddDep(from, to string) {
	fromNode := f.nodeByName(from)
	toNode := f.nodeByName(to)

	// workaround to avoid panic("simple: adding self edge")
	if fromNode == toNode {
		tmpNode := f.nextNode()
		f.graph.SetEdge(simple.Edge{
			F: fromNode,
			T: tmpNode,
		})
		f.graph.SetEdge(simple.Edge{
			F: tmpNode,
			T: toNode,
		})
		return
	}

	f.graph.SetEdge(simple.Edge{
		F: fromNode,
		T: toNode,
	})
}

func (f *graph) CircularDeps() [][]string {
	cycles := topo.DirectedCyclesIn(f.graph)
	sortCycle(cycles)
	r := make([][]string, len(cycles))
	for i, nodeCycle := range cycles {
		r[i] = make([]string, 0, len(nodeCycle))
		for j := 0; j < len(nodeCycle); j++ {
			name, ok := f.nameByNode(simple.Node(nodeCycle[j].ID()))
			if !ok {
				continue
			}
			r[i] = append(r[i], name)
		}
	}
	return r
}

func (f *graph) nodeByName(n string) simple.Node {
	if node, ok := f.namesNodes[n]; ok {
		return node
	}

	node := f.nextNode()
	f.nodesNames[node] = n
	f.namesNodes[n] = node
	return node
}

func (f *graph) nameByNode(node simple.Node) (string, bool) { // workaround to avoid panic("simple: adding self edge")
	name, ok := f.nodesNames[simple.Node(node.ID())]
	return name, ok
}

func (f *graph) nextNode() simple.Node {
	defer func() {
		f.current++
	}()
	return f.current
}

func New() *graph {
	return &graph{
		namesNodes: make(map[string]simple.Node),
		nodesNames: make(map[simple.Node]string),
		graph:      simple.NewDirectedGraph(),
	}
}

// sortCycle sorts the given cycles.
//
// Gonum uses maps internally, so the result of topo.DirectedCyclesIn is unpredictable.
// To make results consistent with each execution, we need to sort them.
func sortCycle(cycles [][]gonumGraph.Node) {
	sort.SliceStable(cycles, func(i, j int) bool {
		l := len(cycles[i])
		if len(cycles[j]) < l {
			l = len(cycles[j])
		}
		for x := 0; x < l; x++ {
			if cycles[i][x] != cycles[j][x] {
				return cycles[i][x].ID() < cycles[j][x].ID()
			}
		}
		return len(cycles[i]) < len(cycles[j])
	})
}
