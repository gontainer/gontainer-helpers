# Container

This package provides a concurrently safe DI container. It supports **scoped variables**, and **atomic hot swapping**.
For bigger projects, it provides a tool for the code generation.

```bash
go get -u github.com/gontainer/gontainer-helpers/v2/container@latest
```

1. [Why?](#why)
2. [Quick start](#quick-start)
3. [Overview](#overview)
   1. [Definitions](#definitions)
   2. [Scopes](#scopes)
   3. [Dependencies](#dependencies)
   4. [Services](#services)
   5. [Parameters](#parameters)
4. [Usage](#usage)
   1. [HotSwap](#hotswap)
   2. [Contextual scope](#contextual-scope)
   3. [Circular dependencies](#circular-dependencies)
   4. [Type conversion](#type-conversion)
   5. [Transactions](#transactions)
   6. [Errors](#errors)
   7. [Examples](#examples)
5. [Code generation](#code-generation)

## Why?

Automatically build and inject scope-aware dependencies.

Let's imagine we work on an endpoint that allows for transferring funds between different accounts.
We have to operate on an SQL transaction, and we have to inject the same transaction into many different objects.

Or maybe you need to reload the configuration without restarting the app? See [hotswap](#hotswap).

<details>
  <summary>See code</summary>

```go
type TransactionHistory struct {
	tx *sql.Tx
}

func (*TransactionHistory) Record(accountID int, amount int) error {
	// TODO
}

type FundsTransferer struct {
	tx *sql.Tx
}

func (*FundsTransferer) Transfer(fromID int, toID int, amount int) error {
	// TODO
}

type Factory interface {
	Tx(context.Context) *sql.Tx
	TransactionHistory(context.Context) *TransactionHistory
	FundsTransferer(context.Context) *FundsTransferer
}

func NewTransferFundsHandler(f Factory) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		/*
			There is an *sql.Tx object here, and two other that rely on *sql.Tx.
			Container creates a new instance of *sql.Tx and injects it to the scope of the current context only.
			You do not need to create a new transaction manually and inject it in many dependencies.
		*/
		var (
			tx         = f.Tx(ctx)
			transferer = f.FundsTransferer(ctx)
			history    = f.TransactionHistory(ctx)
		)

		var err error
		defer func() {
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_ = tx.Rollback()
			}
		}()

		fromID := 1
		toID := 2
		amount := 100

		err = transferer.Transfer(fromID, toID, amount)
		if err != nil {
			return
		}
		err = history.Record(fromID, -amount)
		if err != nil {
			return
		}
		err = history.Record(toID, amount)
		if err != nil {
			return
		}
		_ = tx.Commit()
	})
}
```
</details>

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

---

### Scopes

Scopes are applicable to services only. Parameters are not scope-aware, once created parameter is cached forever,
although `HotSwap` lets overriding them.

1. **Shared** - once created service is cached forever.
2. **Contextual** - service is cached and shared in the current context only.
   The scope is determined by a single invocation of `Get` or `GetTaggedBy`, 
   or by the `context.Context` used in methods `GetInContext` and `GetTaggedByInContext`.
3. **NonShared** - each invocation of a such service will create a new variable.
4. **Default** - a runtime-determining scope.
    If the given service has at least one direct or indirect contextual dependency, 
    its scope will be contextual, otherwise it will be shared.

---

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

**Container**

It refers to the container.

```go
container.NewDependencyContainer()
```

---

### Services

**Creating a new service**

Use either `SetConstructor` or `SetValue`. Constructor MUST be a provider (see [definitions](#definitions)).

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
	s.SetConstructor(func () *Person {
		return &Person{} // it must be a pointer, because `SetName` requires a pointer receiver
	})
	s.AppendCall("SetName", container.NewDependencyValue("Jane"))

	c := container.New()
	c.OverrideService("jane", s)

	jane, _ := c.Get("jane")
	fmt.Printf("%+v\n", jane)
	// Output: &{Name:Jane}
}
```
</details>

**Withers**

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

Use `SetField` or `SetFields`.

<details>
  <summary>See code</summary>

```go
package main

import (
   "fmt"

   "github.com/gontainer/gontainer-helpers/v2/container"
)

type Person struct {
   name string // unexported fields are supported :)
}

func main() {
   s := container.NewService()
   s.SetValue(Person{})
   s.SetField("name", container.NewDependencyParam("name"))

   // alternatively:
   // s.SetFields(map[string]container.Dependency{
   // 	"name": container.NewDependencyValue("name"),
   // })

   c := container.New()
   c.OverrideService("jane", s)
   c.OverrideParam("name", container.NewDependencyValue("Jane"))

   jane, _ := c.Get("jane")
   fmt.Printf("%+v\n", jane)
   // Output: {name:Jane}
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

To define the scope of the given service, use one of the following methods:

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

---

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
		// ...
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
	// ...
})
// decorate your HTTP handler, this function automatically binds contexts with the container
h = container.HTTPHandlerWithContainer(h, c)
```

...or override `server.Handler`...

```go
var (
	s *http.Server
	c *container.Container
)
// your code
// ...
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
			
			// HotSwap guarantees atomicity
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

---

### Contextual scope

DB transactions are the best example to explain the benefits of that feature.
When we have an HTTP server, transaction usually exists in a scope of the given request.
Sometimes we have to inject a transaction into many different structures,
and make sure we won't share it with other HTTP requests.

One approach could be to use `SetTx` methods:

```go
func (r *MyRepository) SetTx(tx *sql.Tx) {
	r.tx = tx
}
```

That solution is error-prone, you can accidentally inject the same instance of `MyRepository` into different requests.
Moreover, when you have many repositories, sometimes even nested ones, injecting `*sql.Tx` manually can be difficult.

Container can solve that problem, you need to instruct it only that the given dependency is contextual, and voila!

<details>
  <summary>See code</summary>

```go
// let's wrap the original container by a custom struct.
// it will let us adding custom getters later
type myContainer struct {
	*container.Container
}

// let's create the constructor for new transactions
func NewTx(db *sql.DB) (*sql.Tx, error) {
	return db.Begin()
}

func buildContainer() *container.Container {
	c := container.New()

	tx := container.NewService()
	tx.SetConstructor(
		NewTx,
		container.NewDependencyService("db"),
	)
	/*
		Here we instruct the container that the scope of tx is contextual.
		By default, the scope of all parent dependencies will be contextual as well.
	*/
	tx.SetScopeContextual()
	c.OverrideService("tx", tx)

	/*
		We will define NewHTTPHandler in the next step,
		now let's register it only in the container
	*/
	myHandler := container.NewService()
	myHandler.SetConstructor(
		NewHTTPHandler,
		container.NewDependencyContainer(),
	)
	myHandler.Tag("http-handler", 0)
	c.OverrideService("myHandler", myHandler)

	// TODO define other dependencies

	/*
		The following code will automatically wrap all http handlers registered to the container
		by func `container.HTTPHandlerWithContainer`.

		It is an equivalent of the following raw code:

			var handler http.Handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// your code here
			})
			handler = decorateHTTPHandler(handler, c)
	*/
	c.AddDecorator(
		"http-handler",
		decorateHTTPHandler,
		container.NewDependencyContainer(),
	)

	return c
}

func decorateHTTPHandler(p container.DecoratorPayload, c *container.Container) http.Handler {
	return container.HTTPHandlerWithContainer(p.Service.(http.Handler), c)
}
```
</details>

Let's build our endpoint now.

<details>
  <summary>See code</summary>

```go
type UserRepository struct {
	tx *sql.Tx
}

type ImageRepository struct {
	tx *sql.Tx
}

// let's define custom getter, they are easier to use
func (c *myContainer) Tx(ctx context.Context) *sql.Tx {
	tx, err := c.GetInContext(ctx, "userRepository")
	// we expect all services to be defined correctly,
	// so we can panic here in case of an error
	if err != nil {
		panic(err)
	}
	return tx.(*sql.Tx)
}

func (c *myContainer) UserRepository(ctx context.Context) *UserRepository {
	u, err := c.GetInContext(ctx, "userRepository")
	if err != nil {
		panic(err)
	}
	return u.(*UserRepository)
}

func (c *myContainer) ImageRepository(ctx context.Context) *ImageRepository {
	i, err := c.GetInContext(ctx, "imageRepository")
	if err != nil {
		panic(err)
	}
	return i.(*ImageRepository)
}

func NewHTTPHandler(c *myContainer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// `tx` refers to the same instance that has been injected to `userRepository` and `imageRepository`
		var (
			tx              = c.Tx(r.Context())
			userRepository  = c.UserRepository(r.Context())
			imageRepository = c.ImageRepository(r.Context())
		)

		var err error
		defer func() {
			// do not forget about committing or rolling back the transaction
			if err == nil {
				tx.Commit()
			} else {
				tx.Rollback()
			}
		}()

		// todo add your logic
	})
}
```
</details>

---

### Circular dependencies

Container automatically detects circular dependencies, and returns a proper error.<br/>
Do not need to worry about `fatal error: stack overflow` :)

<details>
  <summary>See code</summary>

```go
package main

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/v2/container"
)

type Spouse struct {
	Name   string
	Spouse *Spouse
}

func main() {

	wife := container.NewService()
	wife.SetConstructor(func() *Spouse {
		return &Spouse{}
	})
	wife.SetField("Name", container.NewDependencyValue("Mary Jane"))
	wife.SetField("Spouse", container.NewDependencyService("husband"))

	husband := container.NewService()
	husband.SetConstructor(func() *Spouse {
		return &Spouse{}
	})
	husband.SetField("Name", container.NewDependencyValue("Peter Parker"))
	husband.SetField("Spouse", container.NewDependencyService("wife"))

	c := container.New()
	c.OverrideService("wife", wife)
	c.OverrideService("husband", husband)

	_, err := c.Get("wife")
	fmt.Println(err)

	// Output: get("wife"): circular dependencies: @husband -> @wife -> @husband
}
```
</details>

---

### Type conversion

In GO assignments between different types requires explicit type conversion.
Container automatically converts values for more developer-friendly experience.
It supports even a bit more sophisticated conversions of maps and slices, see [copier](../copier/README.md).

<details>
  <summary>See code</summary>

```go
type Superhero struct {
	name string
	age  uint
}

func buildContainer() *container.Container {
	c := container.New()

	ironMan := container.NewService()
	ironMan.SetValue(Superhero{})
	ironMan.SetField("name", container.NewDependencyValue("Tony Stark"))
	// The following value "53" is of the type "int", although we need an "uint"
	// See [Superhero.age]
	// Container automatically converts values for more developer-friendly experience :)
	ironMan.SetField("age", container.NewDependencyValue(53))

	return c
}
```
</details>

---

### Transactions

As an exercise, let's build a small framework that wraps our endpoints and manages transactions automatically.

**ErrorAwareHTTPHandler**

A new type for our endpoints that informs us whether an error has occurred.

<details>
  <summary>See code</summary>

```go
type ErrorAwareHTTPHandler interface {
	Handle(http.ResponseWriter, *http.Request) error
}

type ErrorAwareHTTPHandlerFunc func(http.ResponseWriter, *http.Request) error

func (e ErrorAwareHTTPHandlerFunc) Handle(w http.ResponseWriter, r *http.Request) error {
	return e(w, r)
}
```
</details>

**WrapTransactionHandler**

A small wrapper over the newly created type that implements `http.Handler` interface,
and automatically handles transactions.

<details>
  <summary>See code</summary>

```go
func WrapTransactionHandler(h ErrorAwareHTTPHandler, c *myContainer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// create a copy of the current request, and inject a container aware context
		ctx := container.ContextWithContainer(r.Context(), c)
		r = r.Clone(ctx)

		tx := c.Tx()

		// do not forget about rolling back when your code panics
		defer func() {
			rec := recover()
			if rec != nil {
				_ = tx.Rollback()
				panic(rec)
			}
		}()

		err := h.Handle(w, r)
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	})
}
```
</details>

**Our first endpoint**

<details>
  <summary>See code</summary>

```go
func myHTTPHandler(h http.ResponseWriter, r *http.Request) error {
	// todo some logic
	return errors.New("unexpected error")
}
```
</details>

**The final container**

Let's connect all the dots.

<details>
  <summary>See code</summary>

```go
type myContainer struct {
	*container.Container
}

func (m *myContainer) Tx() *sql.Tx {
	tx, err := m.Get("tx")
	if err != nil {
		panic(err)
	}
	return tx.(*sql.Tx)
}

func buildContainer() *myContainer {
	c := &myContainer{container.New()}

	// decorate all services tagged by "http-errors"
	c.AddDecorator(
		"http-errors",
		func(p container.DecoratorPayload) http.Handler {
			return WrapTransactionHandler(
				p.Service.(ErrorAwareHTTPHandler),
				c,
			)
		},
	)

	// define your error aware http handler
	myHandler := container.NewService()
	myHandler.SetValue(ErrorAwareHTTPHandlerFunc(myHTTPHandler))
	myHandler.Tag("http-errors", 0)
	c.OverrideService("myHandler", myHandler)

	// I assume we may need more endpoints later,
	// let's use the built-in multiplexer [http.ServeMux]
	m := container.NewService()
	m.SetConstructor(http.NewServeMux)
	m.AppendCall(
		"Handle",
		container.NewDependencyValue("/my-error-aware-endpoint"),
		container.NewDependencyService("myHandler"),
	)
	
	return c
}
```
</details>

---

### Errors

This package aims to be as developer-friendly as possible.
To ease the debugging process all errors are as descriptive as possible.
Sometimes you may have more than a single reason why something does not work,
so whenever it is possible you get a multiline error.
Multiline error is a collection of few independent errors.
You can extract them using `grouperror.Collection`, see [grouperror](../grouperror).

<details>
  <summary>See code</summary>

```go
peterParker := container.NewService()
peterParker.SetValue(
	struct {
		name string
	}{},
)
peterParker.SetField("firstname", container.NewDependencyValue("Peter"))
peterParker.SetField("lastname", container.NewDependencyValue("Parker"))

c := container.New()
c.OverrideService("peterParker", peterParker)

_, err := c.Get("peterParker")

fmt.Println(err)

// Output:
// get("peterParker"): set field "firstname": set (*interface {})."firstname": field "firstname" does not exist
// get("peterParker"): set field "lastname": set (*interface {})."lastname": field "lastname" does not exist
```
</details>

---

### Examples

See [examples](examples_test.go).

---

## Code generation

The entire code can be built using a YAML-configuration files.
See [gontainer/gontainer](https://github.com/gontainer/gontainer).