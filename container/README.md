# Container

This package provides a concurrently safe DI container. It supports **scoped variables**, and **atomic hot swapping**.
For bigger projects, it provides a tool for the code generation.

```bash
go get -u github.com/gontainer/gontainer-helpers/v3/container@latest
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
   3. [Transactions](#transactions)
   4. [Circular dependencies](#circular-dependencies)
   5. [Type conversion](#type-conversion)
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

	"github.com/gontainer/gontainer-helpers/v3/container"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/dependency"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/service"
)

type God struct {
	Name string
}

func NewGod(name string) God {
	return God{Name: name}
}

type Gods struct {
	Gods []God
}

func describePoseidon() service.Service {
	p := service.New()
	p.
		SetValue(God{}).
		SetField("Name", dependency.Value("Poseidon")). // field injection
		Tag("olympians", 0)
	return p
}

func describeAthena() service.Service {
	a := service.New()
	a.
		SetValue(God{Name: "Athena"}).
		Tag("olympians", 1) // priority 1 - ladies first :)
	return a
}

func describeZeus() service.Service {
	z := service.New()
	z.
		SetConstructor(NewGod, dependency.Value("Zeus")). // constructor injection
		Tag("olympians", 0)
	return z
}

func describeOlympians() service.Service {
	o := service.New()
	o.
		SetValue(Gods{}).
		SetField("Gods", dependency.Tag("olympians"))
	return o
}

func buildContainer() *container.Container {
	c := container.New()
	c.OverrideServices(service.Services{
		"poseidon":  describePoseidon(),
		"athena":    describeAthena(),
		"zeus":      describeZeus(),
		"olympians": describeOlympians(),
	})

	return c
}

func main() {
	c := buildContainer()
	olympians, _ := c.Get("olympians")
	fmt.Printf("%+v\n", olympians)
	// Output: {Gods:[{Name:Athena} {Name:Poseidon} {Name:Zeus}]}
}
```

## Overview

### Definitions

1. **Service** - any struct, variable, func that you use in your application, e.g. `*sql.DB`.
2. **Parameter** - a variable that holds a configuration. E.g. a password can be a parameter.
3. **Provider** - a function that returns one or two values.
First return may be of any type. Second return if exists must be of a type error.
4. **Factory** - it's a method that is a **provider**.
5. **Wither** - a method that returns a single value always.
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

// or shorter syntax

dependency.Value("https://go.dev/")
```

**Tag**

Search in the container for all services with the given tag.
Sort them by priority first (descending), then by name (alphabetically), and return a slice of them.

```go
container.NewDependencyTag("employee")

// or shorter syntax

dependency.Tag("employee")
```

**Service**

It refers to a service with the given id in the container.

```go
container.NewDependencyService("db")

// or shorter syntax

dependency.Service("db")
```

**Param**

It refers to a param with the given id in the container.

```go
container.NewDependencyParam("db_password")

// or shorter syntax

dependency.Param("db_password")
```

**Provider**

A function that is being invoked whenever the given dependency is requested.

```go
container.NewDependencyProvider(func() string {
    return os.Getenv("DB_PASSWORD")
})

// or shorter syntax

dependency.Provider(func() string {
    return os.Getenv("DB_PASSWORD")
})
```

**Container**

It refers to the container.

```go
container.NewDependencyContainer()

//or shorter syntax

dependency.Container()
```

**Context**

It refers to the current `context.Context`.
If you use `GetInContext` func, the context is the one you passed to that func.
If you use `Get` func, the container uses `context.Background`.

```go
container.NewDependencyContext()

// or shorter syntax

dependency.Context()
```

---

### Services

**Creating a new service**

Use either `SetConstructor` or `SetValue`. Constructor MUST be a provider (see [definitions](#definitions)).

<details>
  <summary>See code</summary>

```go
package main

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/v3/container"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/dependency"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/service"
)

func main() {
	type Person struct {
		Name string
	}

	tonySvc := service.New()
	tonySvc.SetValue(Person{Name: "Tony"}) // create the service using a value

	peterSvc := service.New()
	peterSvc.SetConstructor(
		func(n string) Person { // use a constructor to create a new service
			return Person{
				Name: n,
			}
		},
		dependency.Value("Peter"), // inject the value "Peter" to the constructor
	)

	c := container.New()
	c.OverrideServices(service.Services{
		"tony":  tonySvc,
		"peter": peterSvc,
	})

	tony, _ := c.Get("tony")
	peter, _ := c.Get("peter")
	fmt.Println(tony, peter)

	// Output: {Tony} {Peter}
}
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

	"github.com/gontainer/gontainer-helpers/v3/container"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/dependency"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/service"
)

type Person struct {
	Name string
}

func (p *Person) SetName(n string) {
	p.Name = n
}

func main() {
	s := service.New()
	s.SetConstructor(func() Person {
		// we don't need to use a pointer here, even tho `SetName` requires a pointer receiver :)
		return Person{}
	})
	s.AppendCall("SetName", dependency.Value("Jane"))

	c := container.New()
	c.OverrideService("jane", s)

	jane, _ := c.Get("jane")
	fmt.Printf("%+v\n", jane)
	// Output: {Name:Jane}
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

	"github.com/gontainer/gontainer-helpers/v3/container"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/dependency"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/service"
)

type Person struct {
	Name string
}

func (p Person) WithName(n string) Person {
	p.Name = n
	return p
}

func main() {
	s := service.New()
	s.SetValue(Person{})
	s.AppendWither("WithName", dependency.Value("Jane"))

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

	"github.com/gontainer/gontainer-helpers/v3/container"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/dependency"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/service"
)

type Person struct {
	name string // unexported fields are supported :)
}

func main() {
	s := service.New()
	s.SetValue(Person{})
	s.SetField("name", dependency.Param("name"))

	// alternatively:
	// s.SetFields(field.Fields{
	// 	"name": dependency.Value("name"),
	// })

	c := container.New()
	c.OverrideService("jane", s)
	c.OverrideParam("name", dependency.Value("Jane"))

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
package main

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/v3/container"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/service"
)

func main() {
	type God struct {
		Name string
	}

	type Gods struct {
		Gods []God
	}

	loki := service.New()
	loki.
		SetValue(God{Name: "Loki"}).
		Tag("norse-god", 0) // tag

	thor := service.New()
	thor.
		SetValue(God{Name: "Thor"}).
		Tag("norse-god", 0) // tag

	team := service.New()
	team.
		SetValue(Gods{}).
		SetField("Gods", container.NewDependencyTag("norse-god")) // inject tagged services

	c := container.New()
	c.OverrideServices(service.Services{
		"loki":      loki,
		"thor":      thor,
		"norseGods": team,
	})

	norseGods, _ := c.Get("norseGods")
	fmt.Println(norseGods)

	// Output: {[{Loki} {Thor}]}
}
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
s := service.New()
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

   "github.com/gontainer/gontainer-helpers/v3/container"
   "github.com/gontainer/gontainer-helpers/v3/container/shortcuts/dependency"
)

func main() {
   c := container.New()
   c.OverrideParam("pi", dependency.Value(math.Pi))

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
then blocks attaching other contexts, and other operations on the container,
and modifies the container.

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

...or even easier, use built-in `http.HandlerWithContainer`...

```go
// decorate your HTTP handler, this function automatically binds contexts with the container
myHandler = http.HandlerWithContainer(c, myHandler)
```

...or override `server.Handler`...

```go
s.Handler = http.HandlerWithContainer(c, s.Handler)
```

...or use built-in [`http.ServeMux`](http/http.go)

```go
mux := http.NewServeMux(c)
// use mux.Handle the same way as you use it in the standard library
// or/and
// mux.HandleDynamic(pattern, serviceID)
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
				db := service.New()
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

Container can solve that problem, you need to instruct it only that the given dependency is contextual, and voilà!

<details>
  <summary>See code</summary>

```go
func describeDB() service.Service {
	s := service.New()
	s.SetConstructor(func() (*sql.DB, error) {
		// TODO
	})
	return s
}

func describeTx() service.Service {
	s := service.New()
	// tx, err := db.BeginTx(ctx, nil)
	s.
		SetFactory("db", "BeginTx", dependency.Context(), dependency.Value(nil)).
		SetScopeContextual() // IMPORTANT
	// SetScopeContextual instructs the container to create a new instance of that service for each context
	return s
}

func describeImageRepo() service.Service {
	// ir := repos.ImageRepo{}
	// ir.Tx = c.Get("tx")
	s := service.New()
	s.
		SetValue(repos.ImageRepo{}).
		SetField("Tx", dependency.Service("tx"))
	// NOTE
	// imageRepo has the contextual scope automatically,
	// because it depends on the "tx" service that has the contextual scope
	return s
}

func BuildContainer() *container.Container {
	c := container.New()
	c.OverrideServices(service.Services{
		"db":        describeDB(),
		"tx":        describeTx(),
		"imageRepo": describeImageRepo(),
	})
	
	// TODO define other services

	return c
}
```
</details>

---

### Transactions

As an exercise, let's build a small framework that wraps our endpoints and manages transactions automatically.

**ErrorAwareHandler**

A new type for our endpoints that informs us whether an error has occurred.

<details>
  <summary>See code</summary>

```go
type ErrorAwareHandler interface {
	ServeHTTP(http.ResponseWriter, *http.Request) error
}

type ErrorAwareHandlerFunc func(http.ResponseWriter, *http.Request) error

func (e ErrorAwareHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return e(w, r)
}
```
</details>

**NewAutoCloseTxEndpoint**

A small wrapper over the newly created type. It returns `http.Handler` interface,
and automatically handles transactions.

<details>
  <summary>See code</summary>

```go
func NewAutoCloseTxEndpoint(tx *sql.Tx, handler ErrorAwareHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok := false

		defer func() {
			if !ok {
				// handler panics, so we have to rollback the transaction
				_ = tx.Rollback()
			}
		}()

		err := handler.ServeHTTP(w, r)
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
		ok = true
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

You can find [here](internal/examples/testserver) a sample server written using the above approach.

[See](internal/examples/testserver/http/endpoints.go) how simply the endpoint can look like!
There is no need to create a transaction and inject it into other structs manually.

[See](internal/examples/testserver/container/container.go) how the container is being built.

Execute the following command to start the server:

```
make run-test-server SERVER_ADDR=:8080
```

The server prints in the http response the pointer to the transaction passed to our endpoint
and to its decorator to validate that on different layers in the scope of a single request
we received the same transaction.

**Sample response**

```
Decorator AutoCloseTxEndpoint:
	TxID: 0x1400011e980
MyEndpoint:
	TxID: 0x1400011e980
	userRepo.Tx == imageRepo.Tx: true
```

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

	"github.com/gontainer/gontainer-helpers/v3/container"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/dependency"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/field"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/service"
)

type Spouse struct {
	Name   string
	Spouse *Spouse
}

func main() {
	wife := service.New()
	wife.
		SetConstructor(func() *Spouse {
			return &Spouse{}
		}).
		SetFields(field.Fields{
			"Name":   dependency.Value("Hera"),
			"Spouse": dependency.Service("husband"),
		})

	husband := service.New()
	husband.
		SetConstructor(func() *Spouse {
			return &Spouse{}
		}).
		SetFields(field.Fields{
			"Name":   dependency.Value("Zeus"),
			"Spouse": dependency.Service("wife"),
		})

	c := container.New()
	c.OverrideServices(service.Services{
		"wife":    wife,
		"husband": husband,
	})

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
type Employee struct {
	name string
	age  uint
}

func buildContainer() *container.Container {
	c := container.New()

	jane := service.New()
	jane.SetValue(Employee{})
	jane.SetField("name", container.NewDependencyValue("Jane Doe"))
	// The following value "53" is of the type "int", although we need an "uint"
	// See [Employee.age]
	// Container automatically converts values for more developer-friendly experience :)
	jane.SetField("age", container.NewDependencyValue(53))

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
package main

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/v3/container"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/service"
)

func main() {
	janeDoe := service.New()
	janeDoe.SetValue(
		struct {
			name string
		}{},
	)
	janeDoe.SetField("firstname", container.NewDependencyValue("Jane"))
	janeDoe.SetField("lastname", container.NewDependencyValue("Doe"))

	c := container.New()
	c.OverrideService("janeDoe", janeDoe)

	_, err := c.Get("janeDoe")

	fmt.Println(err)

	// Output:
	// get("janeDoe"): set field "firstname": set (*interface {})."firstname": field "firstname" does not exist
	// get("janeDoe"): set field "lastname": set (*interface {})."lastname": field "lastname" does not exist
}
```
</details>

---

### Examples

See [examples](examples_test.go).

---

## Code generation

The entire code can be built using a YAML-configuration files.
See [gontainer/gontainer](https://github.com/gontainer/gontainer).
