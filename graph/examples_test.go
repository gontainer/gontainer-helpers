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
	"fmt"

	"github.com/gontainer/gontainer-helpers/v3/graph"
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
