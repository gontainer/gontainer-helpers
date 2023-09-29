[![Tests](https://github.com/gontainer/gontainer-helpers/actions/workflows/tests.yml/badge.svg)](https://github.com/gontainer/gontainer-helpers/actions/workflows/tests.yml)
[![Coverage Status](https://coveralls.io/repos/github/gontainer/gontainer-helpers/badge.svg?branch=main)](https://coveralls.io/github/gontainer/gontainer-helpers?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/gontainer/gontainer-helpers)](https://goreportcard.com/report/github.com/gontainer/gontainer-helpers)
[![Go Reference](https://pkg.go.dev/badge/github.com/gontainer/gontainer-helpers.svg)](https://pkg.go.dev/github.com/gontainer/gontainer-helpers)

# Gontainer-helpers

A set of helpers for [Gontainer](https://github.com/gontainer/gontainer).

## Caller

```go
fn := func(a int, b int) int {
    return a * b
}
r, _ := caller.Call(fn, 2, 2)
fmt.Println(r[0])
// Output: 4
```

[More examples](caller/examples_test.go).

## Container

[See examples](container/examples_test.go).

## Copier

```go
var (
    from = 5         // the type of the variable `to` can be different from the type of the variable `from`
    to   interface{} // as long as the `from` is assignable to the `to`
)
err := copier.Copy(from, &to)
fmt.Println(to)
fmt.Println(err)
// Output:
// 5
// <nil>
```

[More examples](copier/examples_test.go).

## Errors

[See examples](errors/examples_test.go).

## Exporter

```go
s, _ := exporter.ToString(false)
fmt.Println(s)
// Output: false
```

[More examples](exporter/examples_test.go).

## Graph

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

[More examples](graph/examples_test.go).

## Setter

```go
person := struct {
    name string
}{}
_ = setter.Set(&person, "name", "Mary")
fmt.Println(person.name)
// Output:
// Mary
```

[More examples](setter/examples_test.go).
