# Graph

This package provides tool to detect circular dependencies and find all dependant nodes in graphs.

```go
g := graph.New()
g.AddDep("company", "tech-team")
g.AddDep("tech-team", "cto")
g.AddDep("cto", "company")
g.AddDep("cto", "ceo")
g.AddDep("ceo", "company")

fmt.Println(g.CircularDeps())

// Output:
// [[company tech-team cto company] [company tech-team cto ceo company]]
```

See [examples](examples_test.go).