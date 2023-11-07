// Copyright (c) 2023 Bartłomiej Krukowski
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

package container_test

import (
	"context"
	"fmt"
	"math"

	"github.com/gontainer/gontainer-helpers/v3/container"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/dependency"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/service"
	"github.com/gontainer/gontainer-helpers/v3/copier"
)

type Person struct {
	Name string
}

func (p *Person) SetName(n string) {
	p.Name = n
}

func (p Person) WithName(n string) Person {
	p.Name = n
	return p
}

type People struct {
	People []Person
}

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

func Example() {
	c := buildContainer()
	olympians, _ := c.Get("olympians")
	fmt.Printf("%+v\n", olympians)
	// Output: {Gods:[{Name:Athena} {Name:Poseidon} {Name:Zeus}]}
}

func ExampleContainer_GetInContext_wrongContext() {
	c := container.New()

	ctx := context.Background()
	// uncomment the following line to remove the panic:
	// ctx = container.ContextWithContainer(ctx)

	five := service.New()
	five.SetValue(5)
	c.OverrideService("five", five)

	defer func() {
		fmt.Println("panic:", recover())
	}()
	_, _ = c.GetInContext(ctx, "five")
	// Output:
	// panic: the given context is not attached to the given container, call `ctx = container.ContextWithContainer(ctx, c)`
}

func ExampleContainer_GetInContext() {
	c := container.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = container.ContextWithContainer(ctx, c)
	nestedCtx, nestedCancel := context.WithCancel(ctx)
	defer nestedCancel()
	anotherCtx, anotherCancel := context.WithCancel(context.Background())
	defer anotherCancel()
	anotherCtx = container.ContextWithContainer(anotherCtx, c)

	pointer := service.New()
	pointer.SetConstructor(func() *int {
		// we use a pointer, so each invocation of the constructor returns a new pointer,
		// we know whether it is a new or cached one by comparing them
		return new(int)
	})
	pointer.SetScopeContextual() // make it contextual!
	c.OverrideService("pointer", pointer)

	var (
		ptrFromGetContext1, _            = c.GetInContext(ctx, "pointer")
		ptrFromGetContext2, _            = c.GetInContext(ctx, "pointer")
		ptrFromGetContextUsingNested, _  = c.GetInContext(nestedCtx, "pointer")
		ptrFromGetContextUsingAnother, _ = c.GetInContext(anotherCtx, "pointer")
		ptrFromGet, _                    = c.Get("pointer")
	)

	fmt.Println(
		"GetInContext() returns the same value for the same context:",
		ptrFromGetContext1 == ptrFromGetContext2,
	)
	fmt.Println(
		"GetInContext() returns the same value for parent and nested one:",
		ptrFromGetContext1 == ptrFromGetContextUsingNested,
	)
	fmt.Println(
		"GetInContext() returns different values for different contexts:",
		ptrFromGetContext1 != ptrFromGetContextUsingAnother,
	)
	fmt.Println(
		"GetInContext() and Get() return different values:",
		ptrFromGetContext1 != ptrFromGet,
	)
	// Output:
	// GetInContext() returns the same value for the same context: true
	// GetInContext() returns the same value for parent and nested one: true
	// GetInContext() returns different values for different contexts: true
	// GetInContext() and Get() return different values: true
}

func ExampleContainer_GetInContext_oneContextManyContainers() {
	c1 := container.New()
	s1 := service.New()
	s1.
		SetValue(5).
		SetScopeContextual()
	c1.OverrideService("number", s1)

	c2 := container.New()
	s2 := service.New()
	s2.
		SetValue(6).
		SetScopeContextual()
	c2.OverrideService("number", s2)

	// attach two containers to the same context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = container.ContextWithContainer(ctx, c1)
	ctx = container.ContextWithContainer(ctx, c2)

	// invoke `GetInContext` to cache the value
	_, _ = c1.GetInContext(ctx, "number")
	_, _ = c2.GetInContext(ctx, "number")

	fmt.Println(c1.GetInContext(ctx, "number"))
	fmt.Println(c2.GetInContext(ctx, "number"))
	// Output:
	// 5 <nil>
	// 6 <nil>
}

func ExampleContainer_Get() {
	// describe "gods"
	gods := service.New()
	gods.
		SetValue(Gods{}).                             // instead of providing a constructor, we can provide a value directly
		SetField("Gods", dependency.Tag("olympians")) // fetch all objects tagged as "olympians", and assign them to the field "Gods"

	// create a Container, and append all services there
	c := container.New()
	c.OverrideServices(service.Services{
		"zeus":   describeZeus(),
		"athena": describeAthena(),
		"gods":   gods,
	})

	godsObject, _ := c.Get("gods")

	fmt.Printf("%+v\n", godsObject)
	// Output: {Gods:[{Name:Athena} {Name:Zeus}]}
}

func ExampleContainer_Get_errorServiceDoesNotExist() {
	type Person struct {
		Name string
	}

	riemann := service.New()
	riemann.SetConstructor(func() Person {
		return Person{}
	})
	riemann.SetField("Name", dependency.Value("Bernhard Riemann"))

	c := container.New()
	// oops... we forgot to add the variable `riemann` to the container
	// c.OverrideService("riemann", riemann)

	_, err := c.Get("riemann")
	fmt.Println(err)
	// Output: get("riemann"): service does not exist
}

func ExampleContainer_Get_errorFieldDoesNotExist() {
	type Person struct {
		Name string
	}

	riemann := service.New()
	riemann.
		SetValue(Person{}).
		SetField("FullName", dependency.Value("Bernhard Riemann")) // it's an invalid field name, it cannot work!

	c := container.New()
	c.OverrideService("riemann", riemann)

	_, err := c.Get("riemann")
	fmt.Println(err)
	// Output:
	// get("riemann"): set field "FullName": set (*interface {})."FullName": field "FullName" does not exist
}

func ExampleContainer_Get_circularDepsServices() {
	type Spouse struct {
		Name   string
		Spouse *Spouse
	}

	wife := service.New()
	wife.
		SetConstructor(func() *Spouse {
			return &Spouse{}
		}).
		SetFields(dependency.Dependencies{
			"Name":   dependency.Value("Hera"),
			"Spouse": dependency.Service("husband"),
		})

	husband := service.New()
	husband.
		SetConstructor(func() *Spouse {
			return &Spouse{}
		}).
		SetFields(dependency.Dependencies{
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

func ExampleContainer_Get_circularDepsParams() {
	c := container.New()

	person := service.New()
	person.
		SetValue(Person{}).
		SetField("name", dependency.Param("name"))

	c.OverrideService("person", person)
	c.OverrideParam("name", dependency.Param("name"))

	_, err := c.Get("person")
	fmt.Println(err)
	// Output: get("person"): field value "name": getParam("name"): circular dependencies: %name% -> %name%
}

func ExampleContainer_CircularDeps() {
	type Spouse struct {
		Name   string
		Spouse *Spouse
	}

	wife := service.New()
	wife.
		SetConstructor(func() *Spouse {
			return &Spouse{}
		}).
		SetFields(dependency.Dependencies{
			"Name":   dependency.Value("Hera"),
			"Spouse": dependency.Service("husband"),
		})

	husband := service.New()
	husband.
		SetConstructor(func() *Spouse {
			return &Spouse{}
		}).
		SetFields(dependency.Dependencies{
			"Name":   dependency.Value("Zeus"),
			"Spouse": dependency.Service("wife"),
		})

	c := container.New()
	c.OverrideServices(service.Services{
		"wife":    wife,
		"husband": husband,
	})
	c.OverrideParam("name", dependency.Param("name"))

	fmt.Println(c.CircularDeps())
	// Output:
	// CircularDeps(): @husband -> @wife -> @husband
	// CircularDeps(): %name% -> %name%
}

func ExampleContainer_Get_setterInjection() {
	riemannSvc := service.New()
	riemannSvc.
		SetConstructor(func() Person { // we don't need to use a pointer here, even tho `SetName` requires a pointer receiver :)
			return Person{}
		}).
		AppendCall("SetName", dependency.Value("Bernhard Riemann"))

	c := container.New()
	c.OverrideService("riemann", riemannSvc)

	riemann, _ := c.Get("riemann")
	fmt.Println(riemann)
	// Output: {Bernhard Riemann}
}

func ExampleContainer_Get_witherInjection() {
	riemannSvc := service.New()
	riemannSvc.
		SetConstructor(func() Person {
			return Person{}
		}).
		AppendWither("WithName", dependency.Value("Bernhard Riemann"))

	c := container.New()
	c.OverrideService("riemann", riemannSvc)

	riemann, _ := c.Get("riemann")
	fmt.Println(riemann)
	// Output: {Bernhard Riemann}
}

func ExampleContainer_Get_fieldInjection() {
	riemannSvc := service.New()
	riemannSvc.
		SetConstructor(func() Person {
			return Person{}
		}).
		SetField("Name", dependency.Value("Bernhard Riemann"))

	c := container.New()
	c.OverrideService("riemann", riemannSvc)

	riemann, _ := c.Get("riemann")
	fmt.Println(riemann)
	// Output: {Bernhard Riemann}
}

type Greeter interface {
	Greet() string
}

type greeter struct{}

func (g greeter) Greet() string {
	return "How are you?"
}

type politeGreeter struct {
	parent Greeter
}

func (p politeGreeter) Greet() string {
	return fmt.Sprintf("Hello! %s", p.parent.Greet())
}

func PoliteGreeterDecorator(payload container.DecoratorPayload) politeGreeter {
	return politeGreeter{
		parent: payload.Service.(greeter),
	}
}

func ExampleContainer_AddDecorator() {
	g := service.New()
	g.
		SetValue(greeter{}).
		Tag("greeter-tag", 0)

	c := container.New()
	c.OverrideService("greeter", g)
	c.AddDecorator("greeter-tag", PoliteGreeterDecorator)

	var greeterObj Greeter
	tmp, _ := c.Get("greeter")
	_ = copier.Copy(tmp, &greeterObj, true)
	fmt.Println(greeterObj.Greet())
	// Output: Hello! How are you?
}

func ExampleContainer_Get_scopeShared() {
	i := 0

	num := service.New()
	num.
		SetConstructor(func() int {
			i++ // each invocation increments the value
			return i
		}).
		SetScopeShared()

	c := container.New()
	c.OverrideService("number", num)

	first, _ := c.Get("number")
	second, _ := c.Get("number")

	// first is equal to second, because the scope is shared
	fmt.Println(first, second)
	// Output: 1 1
}

func ExampleContainer_Get_scopeNonShared() {
	i := 0

	num := service.New()
	num.
		SetConstructor(func() int {
			i++
			return i
		}).
		SetScopeNonShared()

	c := container.New()
	c.OverrideService("number", num)

	first, _ := c.Get("number")
	second, _ := c.Get("number")

	// first is not equal to second, because the scope is private
	fmt.Println(first, second)
	// Output: 1 2
}

func ExampleContainer_Get_taggedServices() {
	c := buildContainer()
	olympians, _ := c.Get("olympians")
	fmt.Printf("%+v\n", olympians)
	// Output: {Gods:[{Name:Athena} {Name:Poseidon} {Name:Zeus}]}
}

func ExampleContainer_GetTaggedBy() {
	type Person struct {
		Name string
	}

	p1 := service.New()
	p1.
		SetValue(Person{}).
		SetField("Name", dependency.Value("person1")).
		Tag("person", 0) // priority 0

	p2 := service.New()
	p2.
		SetValue(Person{}).
		SetField("Name", dependency.Value("person2")).
		Tag("person", 1) // priority 1

	p3 := service.New()
	p3.
		SetValue(Person{}).
		SetField("Name", dependency.Value("person3")).
		Tag("person", 1) // priority 1

	c := container.New()
	c.OverrideServices(service.Services{
		"p1": p1,
		"p2": p2,
		"p3": p3,
	})

	// it returns all services tagged by "person", sorted by the tag priority
	people, _ := c.GetTaggedBy("person")

	fmt.Println(people)
	// Output:
	// [{person2} {person3} {person1}]
}

func ExampleContainer_Get_invalidConstructorParameters() {
	type Server struct {
		Host string
		Port int
	}

	s := service.New()
	// invalid arguments
	s.SetConstructor(
		func(host string, port int) *Server {
			return &Server{
				Host: host,
				Port: port,
			}
		},
		dependency.Value(nil),         // it should be a string!
		dependency.Value("localhost"), // it should be an int!
	)

	c := container.New()
	c.OverrideService("server", s)

	_, err := c.Get("server")
	fmt.Println(err)
	// Output:
	// get("server"): constructor: cannot call provider func(string, int) *container_test.Server: arg0: cannot convert <nil> to string
	// get("server"): constructor: cannot call provider func(string, int) *container_test.Server: arg1: cannot convert string to int
}

func ExampleContainer_IsTaggedBy() {
	c := container.New()

	pi := service.New()
	pi.SetValue(math.Pi)
	c.OverrideService("pi", pi)

	three := service.New()
	three.
		SetValue(3).
		Tag("int", 0)
	c.OverrideService("three", three)

	fmt.Printf("pi is tagged by int: %v\n", c.IsTaggedBy("pi", "int"))
	fmt.Printf("three is tagged by int: %v\n", c.IsTaggedBy("three", "int"))
	// Output:
	// pi is tagged by int: false
	// three is tagged by int: true
}

func ExampleContainer_IsTaggedBy_serviceDoesNotExist() {
	c := container.New()
	fmt.Println(c.IsTaggedBy("service", "tag"))
	// Output: false
}

func ExampleNewDependencyProvider() {
	c := container.New()
	c.OverrideParam("pi", dependency.Provider(func() float64 {
		return math.Pi
	}))
	pi, _ := c.GetParam("pi")
	fmt.Printf("%0.2f\n", pi)
	// Output: 3.14
}
