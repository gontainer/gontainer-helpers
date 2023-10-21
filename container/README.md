# Container

This package provides a concurrently safe DI container. It supports scoped variables, and hot swapping.

## Quick start

```go
package main

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/container"
)

type Superhero struct {
	Name string
}

func NewSuperhero(name string) Superhero {
	return Superhero{Name: name}
}

type Team struct {
	Superheroes []Superhero
}

func buildContainer() *container.Container {
	// describe Iron Man
	ironMan := container.NewService()
	ironMan.SetValue(Superhero{})
	ironMan.SetField("Name", container.NewDependencyValue("Iron Man"))
	ironMan.Tag("avengers", 0)

	// describe Thor
	thor := container.NewService()
	thor.SetValue(Superhero{
		Name: "Thor",
	})
	thor.Tag("avengers", 1) // Thor has higher priority

	// describe Hulk
	hulk := container.NewService()
	hulk.SetConstructor(
		NewSuperhero,
		container.NewDependencyValue("Hulk"),
	)
	hulk.Tag("avengers", 0)

	// describe Avengers
	avengers := container.NewService()
	avengers.SetValue(Team{})
	avengers.SetField("Superheroes", container.NewDependencyTag("avengers"))

	c := container.New()
	c.OverrideService("ironMan", ironMan)
	c.OverrideService("thor", thor)
	c.OverrideService("hulk", hulk)
	c.OverrideService("avengers", avengers)

	return c
}

func main() {
	c := buildContainer()
	avengers, _ := c.Get("avengers")
	fmt.Printf("%+v\n", avengers)
	// Output: {Superheroes:[{Name:Thor} {Name:Hulk} {Name:Iron Man}]}
}
```

## Definitions

1. **Service** - any struct, variable, func that you use in your application, e.g. `*sql.DB`.
2. **Parameter** - a variable that holds a configuration. E.g. a password can be a parameter.
3. **Provider** - a function that returns one or two values.
First return may be of any type. Second return if exists must be of a type error.
4. **Wither** - a method that returns a single value always.
Withers in opposition to setters are being used to achieve immutable structures.

## Scopes

Scopes are applicable to services only. Parameters are not scope-aware, once created parameter is cached forever,
although `HotSwap` lets overriding them.

1. **Shared** - once created service is cached forever.
2. **Contextual** - service is cached and shared in the current context only.
3. **NonShared** - each invocation of a such service will create a new variable.
4. **Default** - a runtime-determining scope.
If the given service has at least one direct or indirect contextual dependency,
its scope will be contextual, otherwise it will be shared.
