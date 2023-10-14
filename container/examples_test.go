package container_test

import (
	"context"
	"fmt"
	"math"

	"github.com/gontainer/gontainer-helpers/container"
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

func ExampleNewContainer_wrongContext() {
	c := container.NewContainer()

	ctx := context.Background()
	// uncomment the following line to remove the panic:
	// ctx = c.ContextWithContainer(ctx)

	five := container.NewService()
	five.SetValue(5)
	c.OverrideService("five", five)

	defer func() {
		fmt.Println("panic:", recover())
	}()
	_, _ = c.GetContextual(ctx, "five")

	// Output:
	// panic: the given context is not attached to the given container, call `ctx = c.ContextWithContainer(ctx)`
}

func ExampleNewContainer_getContextual() {
	c := container.NewContainer()

	ctx := context.Background()
	ctx = c.ContextWithContainer(ctx)
	nestedCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	anotherCtx := c.ContextWithContainer(context.Background())

	pointer := container.NewService()
	pointer.SetConstructor(func() *int {
		// we use a pointer, so each invocation of the constructor returns a new pointer,
		// we know whether it is a new or cached one by comparing them
		return new(int)
	})
	pointer.ScopeContextual() // make it contextual!
	c.OverrideService("pointer", pointer)

	var (
		ptrFromGetContext1, _            = c.GetContextual(ctx, "pointer")
		ptrFromGetContext2, _            = c.GetContextual(ctx, "pointer")
		ptrFromGetContextUsingNested, _  = c.GetContextual(nestedCtx, "pointer")
		ptrFromGetContextUsingAnother, _ = c.GetContextual(anotherCtx, "pointer")
		ptrFromGet, _                    = c.Get("pointer")
	)

	fmt.Println(
		"GetContextual() returns the same value for the same context:",
		ptrFromGetContext1 == ptrFromGetContext2,
	)
	fmt.Println(
		"GetContextual() returns the same value for parent and nested one:",
		ptrFromGetContext1 == ptrFromGetContextUsingNested,
	)
	fmt.Println(
		"GetContextual() returns different values for different contexts:",
		ptrFromGetContext1 != ptrFromGetContextUsingAnother,
	)
	fmt.Println(
		"GetContextual() and Get() return different values:",
		ptrFromGetContext1 != ptrFromGet,
	)

	// Output:
	// GetContextual() returns the same value for the same context: true
	// GetContextual() returns the same value for parent and nested one: true
	// GetContextual() returns different values for different contexts: true
	// GetContextual() and Get() return different values: true
}

func ExampleNewContainer_simple() {
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

func ExampleNewContainer_errorServiceDoesNotExist() {
	mary := container.NewService()
	mary.SetConstructor(func() Person {
		return Person{}
	})
	mary.SetField("Name", container.NewDependencyValue("Mary Jane"))

	c := container.NewContainer()
	// oops... we forgot to add the variable `mary` to the container
	// c.OverrideService("mary", mary)

	_, err := c.Get("mary")
	fmt.Println(err)

	// Output: container.get("mary"): service does not exist
}

func ExampleNewContainer_errorFieldDoesNotExist() {
	mary := container.NewService()
	mary.SetConstructor(func() Person {
		return Person{}
	})
	// it's an invalid field name, it cannot work!
	mary.SetField("FullName", container.NewDependencyValue("Mary Jane"))

	c := container.NewContainer()
	c.OverrideService("mary", mary)

	_, err := c.Get("mary")
	fmt.Println(err)

	// Output:
	// container.get("mary"): set field "FullName": set `*interface {}`."FullName": field `FullName` does not exist
}

type Spouse struct {
	Name   string
	Spouse *Spouse
}

func ExampleNewContainer_circularDependency() {
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

	c := container.NewContainer()
	c.OverrideService("wife", wife)
	c.OverrideService("husband", husband)

	_, err := c.Get("wife")
	fmt.Println(err)

	// Output: container.get("wife"): circular dependencies: @husband -> @wife -> @husband
}

func ExampleNewContainer_setter() {
	mary := container.NewService()
	mary.SetConstructor(func() *Person { // we have to use a pointer, because we use a setter
		return &Person{}
	})
	mary.AppendCall("SetName", container.NewDependencyValue("Mary Jane"))

	c := container.NewContainer()
	c.OverrideService("mary", mary)

	maryObject, _ := c.Get("mary")
	fmt.Println(maryObject)

	// Output: &{Mary Jane}
}

func ExampleNewContainer_wither() {
	mary := container.NewService()
	mary.SetConstructor(func() Person {
		return Person{}
	})
	mary.AppendWither("WithName", container.NewDependencyValue("Mary Jane"))

	c := container.NewContainer()
	c.OverrideService("mary", mary)

	maryObject, _ := c.Get("mary")
	fmt.Println(maryObject)

	// Output: {Mary Jane}
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

func ExampleNewContainer_decorator() {
	g := container.NewService()
	g.SetValue(greeter{})
	g.Tag("greeter-tag", 0)

	c := container.NewContainer()
	c.OverrideService("greeter", g)
	c.AddDecorator("greeter-tag", PoliteGreeterDecorator)

	var greeterObj Greeter
	_ = c.CopyServiceTo("greeter", &greeterObj)
	fmt.Println(greeterObj.Greet())

	// Output: Hello! How are you?
}

func ExampleNewContainer_scopeShared() {
	i := 0

	num := container.NewService()
	num.SetConstructor(func() int {
		i++
		return i
	})
	num.ScopeShared()

	c := container.NewContainer()
	c.OverrideService("number", num)

	first, _ := c.Get("number")
	second, _ := c.Get("number")

	// first is equal to second, because the scope is shared
	fmt.Println(first, second)

	// Output: 1 1
}

func ExampleNewContainer_scopeNonShared() {
	i := 0

	num := container.NewService()
	num.SetConstructor(func() int {
		i++
		return i
	})
	num.ScopeNonShared()

	c := container.NewContainer()
	c.OverrideService("number", num)

	first, _ := c.Get("number")
	second, _ := c.Get("number")

	// first is not equal to second, because the scope is private
	fmt.Println(first, second)

	// Output: 1 2
}

func ExampleNewContainer_copyServiceToOK() {
	p := container.NewService()
	p.SetValue(Person{})
	p.SetField("Name", container.NewDependencyValue("Mary Jane"))

	c := container.NewContainer()
	c.OverrideService("mary", p)

	var mary Person
	_ = c.CopyServiceTo("mary", &mary)
	fmt.Println(mary)

	// Output: {Mary Jane}
}

func ExampleNewContainer_copyServiceToError() {
	p := container.NewService()
	p.SetValue(Person{})
	p.SetField("Name", container.NewDependencyValue("Mary Jane"))

	c := container.NewContainer()
	c.OverrideService("mary", p)

	var mary People
	err := c.CopyServiceTo("mary", &mary)
	fmt.Println(err)

	// Output:
	// container.CopyServiceTo("mary"): reflect.Set: value of type container_test.Person is not assignable to type container_test.People
}

func ExampleNewContainer_getTaggedBy() {
	p1 := container.NewService()
	p1.SetValue(Person{})
	p1.SetField("Name", container.NewDependencyValue("person1"))
	p1.Tag("person", 0) // priority 0

	p2 := container.NewService()
	p2.SetValue(Person{})
	p2.SetField("Name", container.NewDependencyValue("person2"))
	p2.Tag("person", 1) // priority 1

	p3 := container.NewService()
	p3.SetValue(Person{})
	p3.SetField("Name", container.NewDependencyValue("person3"))
	p3.Tag("person", 1) // priority 1

	c := container.NewContainer()
	c.OverrideService("p1", p1)
	c.OverrideService("p2", p2)
	c.OverrideService("p3", p3)

	// it returns all services tagged by "person", sorted by the tag priority
	people, _ := c.GetTaggedBy("person")

	fmt.Println(people)

	// Output:
	// [{person2} {person3} {person1}]
}

type Server struct {
	Host string
	Port int
}

func NewServer(host string, port int) *Server {
	return &Server{Host: host, Port: port}
}

func ExampleNewContainer_invalidConstructorParameters() {
	s := container.NewService()
	// invalid arguments
	s.SetConstructor(
		NewServer,
		container.NewDependencyValue(nil),         // it should be a string!
		container.NewDependencyValue("localhost"), // it should be an int!
	)

	c := container.NewContainer()
	c.OverrideService("server", s)

	_, err := c.Get("server")
	fmt.Println(err)

	// Output:
	// container.get("server"): constructor: arg0: cannot cast `<nil>` to `string`
	// container.get("server"): constructor: arg1: cannot cast `string` to `int`
}

func ExampleNewContainer_isTaggedBy() {
	c := container.NewContainer()

	pi := container.NewService()
	pi.SetValue(math.Pi)
	c.OverrideService("pi", pi)

	three := container.NewService()
	three.SetValue(3)
	three.Tag("int", 0)
	c.OverrideService("three", three)

	fmt.Println(
		c.IsTaggedBy("pi", "int"),
		c.IsTaggedBy("three", "int"),
	)

	// Output:
	// false true
}
