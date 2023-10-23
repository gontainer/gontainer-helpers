# Container

This package provides a concurrently safe DI container. It supports scoped variables, and hot swapping.

```bash
go get -u github.com/gontainer/gontainer-helpers/v2/container@latest
```

1. [Quick start](#quick-start)
2. [Overview](#overview)
   1. [Definitions](#definitions)
   2. [Scopes](#scopes)
   3. [Dependencies](#dependencies)
   4. [Services](#services)
   5. [Parameters](#parameters)
3. [Usage](#usage)
   1. [HotSwap](#hotswap)
   2. [Contextual scope](#contextual-scope)

## Quick start

```go
package main

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/v2/container"
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
	thor.Tag("avengers", 1) // Thor has a higher priority

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

## Overview

### Definitions

1. **Service** - any struct, variable, func that you use in your application, e.g. `*sql.DB`.
2. **Parameter** - a variable that holds a configuration. E.g. a password can be a parameter.
3. **Provider** - a function that returns one or two values.
First return may be of any type. Second return if exists must be of a type error.
4. **Wither** - a method that returns a single value always.
Withers in opposition to setters are being used to achieve immutable structures.

**Sample providers**

<details>
  <summary>See code</summary>

```go
func GetPassword() string {
	return os.Getenv("PASSWORD")
}
```

```go
func NewDB(username, password string) (*sql.DB, error) {
	return sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/test", username, password))
}
```
</details>

**Sample wither**

<details>
  <summary>See code</summary>

```go
type Person struct {
	Name string
}

// WithName is a wither.
func (p Person) WithName(n string) Person {
	p.Name = n
	return p
}
```
</details>

### Scopes

Scopes are applicable to services only. Parameters are not scope-aware, once created parameter is cached forever,
although `HotSwap` lets overriding them.

1. **Shared** - once created service is cached forever.
2. **Contextual** - service is cached and shared in the current context only.
3. **NonShared** - each invocation of a such service will create a new variable.
4. **Default** - a runtime-determining scope.
If the given service has at least one direct or indirect contextual dependency,
its scope will be contextual, otherwise it will be shared.

### Dependencies

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

### Services

**Creating a new service**

Use either `SetConstructor` or `SetValue`. Constructor is a provider (see [definitions](#definitions)).

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

	"github.com/gontainer/gontainer-helpers/v2/container"
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

Use `AppendWither`.

<details>
  <summary>See code</summary>

```go
package main

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/v2/container"
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

   "github.com/gontainer/gontainer-helpers/v2/container"
)

type Person struct {
   Name string
}

func main() {
   s := container.NewService()
   s.SetValue(Person{})
   s.SetField("Name", container.NewDependencyParam("name"))

   c := container.New()
   c.OverrideService("jane", s)
   c.OverrideParam("name", container.NewDependencyValue("Jane"))

   jane, _ := c.Get("jane")
   fmt.Printf("%+v\n", jane)
   // Output: {Name:Jane}
}
```
</details>

**Tagging**

To tag a service use the function `Tag`. The first argument is a tag name, the second one is a priority,
used for sorting services, whenever the given tag is requested.

<details>
  <summary>See code</summary>

```go
s := container.NewService()
s.Tag("handler", 0)
```
</details>

**Scope**

To define the scope of the given service, use on of the following methods:

1. `SetScopeDefault`
2. `SetScopeShared`
3. `SetScopeContextual`
4. `SetScopeNonShared`

<details>
  <summary>See code</summary>

```go
s := container.NewService()
s.SetScopeContextual()
```
</details>

### Parameters

Parameters are being registered as dependencies. See [dependencies](#dependencies).

<details>
  <summary>See code</summary>

```go
package main

import (
	"fmt"
	"math"

	"github.com/gontainer/gontainer-helpers/v2/container"
)

func main() {
	c := container.New()
	c.OverrideParam("pi", container.NewDependencyValue(math.Pi))

	pi, _ := c.GetParam("pi")
	fmt.Printf("%.2f\n", pi)
	// Output: 3.14
}
```
</details>

## Usage

### HotSwap

HotSwap lets us gracefully change anything in the container in real time.
It waits till all contexts attached to the container are done,
then blocks attaching other contexts, and modifies the container.

Let's create an HTTP handler that will use the container:

<details>
  <summary>See code</summary>

```go
func MyHTTPEndpoint(c *container.Container) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// it creates a new context and guarantees
		// that the container won't be modified during that request
		// if you use the HotSwap function
		ctx := container.ContextWithContainer(r.Context(), c)
		r = r.Clone(ctx)

		// your code
	})
}
```

...or even easier, use built-in `HTTPHandlerWithContainer`...

```go
var (
	h http.Handler
	c *container.Container
)

// create your HTTP handler
h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// your code
})
// decorate your HTTP handler, this function automatically binds requests' contexts with the container
h = container.HTTPHandlerWithContainer(h, c)
```

...or override `server.Handler`...

```go
var (
	s *http.Server
	c *container.Container
)
// your code
s.Handler = container.HTTPHandlerWithContainer(s.Handler, c)
```
</details>

Now we need to change the configuration.
Instead of restarting the server, we can use the HotSwap function.
E.g.:

<details>
  <summary>See code</summary>

```go
// RefreshConfigEveryMinute refreshes the configuration of the container every minute
func RefreshConfigEveryMinute(c *container.Container) {
	go func () {
		for {
			<-time.After(time.Minute)
			
			c.HotSwap(func(c container.MutableContainer) {
				// override the value of a param
				// the cache for that param is automatically invalidated
				c.OverrideParam("my-param", container.NewDependencyValue(125))

				// override a service
				// the cache for that service is automatically invalidated
				db := container.NewService()
				db.SetConstructor(
					sql.Open,
					container.NewDependencyValue("mysql"),
					container.NewDependencyParam("dataSourceName"),
				)

				// invalidate the cache for the given params...
				c.InvalidateParamsCache("paramA", "paramB")
				// ... or for all of them
				c.InvalidateAllParamsCache()

				// invalidate the cache for the given service...
				c.InvalidateServicesCache("serviceA", "serviceB")
				/// or for all of them
				c.InvalidateAllServicesCache()
			})
		}
	}()
}
```
</details>

### Contextual scope

TODO
