package graph_test

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/v2/graph"
)

func ExampleNew_circularDeps() {
	g := graph.New()
	g.AddDep("company", "tech-team")
	g.AddDep("tech-team", "cto")
	g.AddDep("cto", "company")
	g.AddDep("cto", "ceo")
	g.AddDep("ceo", "company")

	fmt.Println(g.CircularDeps())

	// Output:
	// [[company tech-team cto company] [company tech-team cto ceo company]]
}

func ExampleNew_deps() {
	g := graph.New()
	g.AddDep("company", "tech-team")
	g.AddDep("tech-team", "cto")

	fmt.Println(g.Deps("company"))

	// Output:
	// [cto tech-team]
}
