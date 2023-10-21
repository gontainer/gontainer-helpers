# Container

This package provides a concurrently safe DI container. It supports scoped variables, and hot swapping.

1. [Quick start](#quick-start)
2. [Definitions](#definitions)
3. [Scopes](#scopes)
4. [Dependencies](#dependencies)
5. [Services](#services)

```bash
go get -u github.com/gontainer/gontainer-helpers/container@latest
```

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

## Dependencies

Dependencies describe values we inject to our services.

**Value**

Hardcoded value. The simplest possible dependency.

```go
container.NewDependencyValue("https://go.dev/")
```

**Tag**

Search in the container for all services with the given tag.
Sort them by priority first, then by name, and return a slice of them.

```go
container.NewDependencyTag("employee")
```

**Service**

It refers to a service with the given id in the container.

```go
container.NewDependencyService("db")
```

**Param**

It refers to a param with the given id in the container.

```go
container.NewDependencyParam("db.password")
```

**Provider**

A function that is being invoked whenever the given dependency is requested.

```go
container.NewDependencyProvider(func() string {
    return os.Getenv("DB_PASSWORD")
})
```

## Services

**Creating a new service**

Use `SetConstructor`.

<details>
  <summary>See code</summary>

```go
type Person struct {
	Name string
}

svc1 := container.NewService()
svc1.SetValue(Person{}) // create the service using a value

svc2 := container.NewService()
svc2.SetConstructor(
    func (n string) Person { // use a constructor to create a new service
        return Person{
            Name: n
        }       
    },
    container.NewDependencyParam("name"), // inject parameter "name" to the constructor
)
```
</details>

**Setter injection**

Use `AppendCall`.

<details>
  <summary>See code</summary>

```go
package main

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/container"
)

type Person struct {
	Name string
}

func (p *Person) SetName(n string) {
	p.Name = n
}

func main() {
	s := container.NewService()
	s.SetValue(&Person{}) // it must be a pointer
	s.AppendCall("SetName", container.NewDependencyValue("Jane"))

	c := container.New()
	c.OverrideService("jane", s)

	jane, _ := c.Get("jane")
	fmt.Printf("%+v\n", jane)
	// Output: &{Name:Jane}
}
```
</details>

**Wither injection**

Use `AppendCall`.

<details>
  <summary>See code</summary>

```go
package main

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/container"
)

type Person struct {
	Name string
}

func (p Person) WithName(n string) Person {
	p.Name = n
	return p
}

func main() {
	s := container.NewService()
	s.SetValue(Person{})
	s.AppendWither("WithName", container.NewDependencyValue("Jane"))

	c := container.New()
	c.OverrideService("jane", s)

	jane, _ := c.Get("jane")
	fmt.Printf("%+v\n", jane)
	// Output: {Name:Jane}
}
```
</details>

**Field injection**

Use `SetField`.

<details>
  <summary>See code</summary>

```go
package main

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/container"
)

type Person struct {
	Name string
}

func main() {
	s := container.NewService()
	s.SetValue(Person{})
	s.SetField("Name", container.NewDependencyValue("Jane"))

	c := container.New()
	c.OverrideService("jane", s)

	jane, _ := c.Get("jane")
	fmt.Printf("%+v\n", jane)
	// Output: {Name:Jane}
}
```
</details>

**Tagging**

TODO
