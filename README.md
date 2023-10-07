[![Tests](https://github.com/gontainer/gontainer-helpers/actions/workflows/tests.yml/badge.svg)](https://github.com/gontainer/gontainer-helpers/actions/workflows/tests.yml)
[![Coverage Status](https://coveralls.io/repos/github/gontainer/gontainer-helpers/badge.svg?branch=main)](https://coveralls.io/github/gontainer/gontainer-helpers?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/gontainer/gontainer-helpers)](https://goreportcard.com/report/github.com/gontainer/gontainer-helpers)
[![Go Reference](https://pkg.go.dev/badge/github.com/gontainer/gontainer-helpers.svg)](https://pkg.go.dev/github.com/gontainer/gontainer-helpers)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=gontainer_gontainer-helpers&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=gontainer_gontainer-helpers)

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

[More examples](caller/examples_test.go)

## Container

Provides a concurrent-safe DI container.

```go
package main

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/container"
)

type Person struct {
	Name string
}

type People struct {
    People []Person
}

func main()  {
	// create Mary Jane
	mary := container.NewService()
	mary.SetConstructor(func() Person {
		return Person{}
	})
	mary.SetField("Name", container.NewDependencyValue("Mary Jane"))
	mary.Tag("person", 1) // priority = 1, ladies first :)

	// create Peter Parker
	peter := container.NewService()
	peter.SetConstructor(func() Person {
		return Person{}
	})
	peter.SetField("Name", container.NewDependencyProvider(func() string {
		return "Peter Parker"
	}))
	peter.Tag("person", 0)

	// create "people"
	people := container.NewService()
	people.SetValue(People{})                                       // instead of providing a constructor, we can provide a value directly
	people.SetField("People", container.NewDependencyTag("person")) // fetch all objects tagged as "person", and assign them to the field "people"

	// create a container, and append all services there
	c := container.NewContainer()
	c.OverrideService("mary", mary)
	c.OverrideService("peter", peter)
	c.OverrideService("people", people)

	// instead of these 2 following lines,
	// you can write:
	//
	// peopleObject, _ := c.Get("people")
	var peopleObject People
	_ = c.CopyServiceTo("people", &peopleObject)

	fmt.Printf("%+v\n", peopleObject)

	// Output: {People:[{Name:Mary Jane} {Name:Peter Parker}]}
}
```

[More examples](container/examples_test.go)

## Copier

```go
var (
    from = 5         // the type of the variable `to` can be different from the type of the variable `from`
    to   interface{} // as long as the underlying value of `from` is assignable to the `to`
)
_ = copier.Copy(from, &to)
fmt.Println(to)
// Output:
// 5
```

[More examples](copier/examples_test.go)

## Errors

[See examples](errors/examples_test.go)

## Exporter

```go
s, _ := exporter.Export(float32(3.1416))
fmt.Println(s)
// Output: float32(3.1416)
```

[More examples](exporter/examples_test.go)

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

[More examples](graph/examples_test.go)

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

[More examples](setter/examples_test.go)
