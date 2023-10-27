package container_test

import (
	"context"
	"fmt"
	"math"

	"github.com/gontainer/gontainer-helpers/v2/container"
	"github.com/gontainer/gontainer-helpers/v2/copier"
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

func ExampleContainer_GetInContext_wrongContext() {
	c := container.New()

	ctx := context.Background()
	// uncomment the following line to remove the panic:
	// ctx = container.ContextWithContainer(ctx)

	five := container.NewService()
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

	pointer := container.NewService()
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
	s1 := container.NewService()
	s1.SetValue(5)
	s1.SetScopeContextual()
	c1.OverrideService("number", s1)

	c2 := container.New()
	s2 := container.NewService()
	s2.SetValue(6)
	s2.SetScopeContextual()
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
	type Person struct {
		Name string
	}

	type People struct {
		People []Person
	}

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

	// create a Container, and append all services there
	c := container.New()
	c.OverrideService("mary", mary)
	c.OverrideService("peter", peter)
	c.OverrideService("people", people)

	peopleObject, _ := c.Get("people")

	fmt.Printf("%+v\n", peopleObject)

	// Output: {People:[{Name:Mary Jane} {Name:Peter Parker}]}
}

func ExampleContainer_Get_errorServiceDoesNotExist() {
	type Person struct {
		Name string
	}

	mary := container.NewService()
	mary.SetConstructor(func() Person {
		return Person{}
	})
	mary.SetField("Name", container.NewDependencyValue("Mary Jane"))

	c := container.New()
	// oops... we forgot to add the variable `mary` to the Container
	// c.OverrideService("mary", mary)

	_, err := c.Get("mary")
	fmt.Println(err)

	// Output: get("mary"): service does not exist
}

func ExampleContainer_Get_errorFieldDoesNotExist() {
	type Person struct {
		Name string
	}

	mary := container.NewService()
	mary.SetConstructor(func() Person {
		return Person{}
	})
	// it's an invalid field name, it cannot work!
	mary.SetField("FullName", container.NewDependencyValue("Mary Jane"))

	c := container.New()
	c.OverrideService("mary", mary)

	_, err := c.Get("mary")
	fmt.Println(err)

	// Output:
	// get("mary"): set field "FullName": set (*interface {})."FullName": field "FullName" does not exist
}

func ExampleContainer_Get_circularDepsServices() {
	type Spouse struct {
		Name   string
		Spouse *Spouse
	}

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

func ExampleContainer_Get_circularDepsParams() {
	c := container.New()

	person := container.NewService()
	person.SetValue(Person{})
	person.SetField("name", container.NewDependencyParam("name"))

	c.OverrideService("person", person)
	c.OverrideParam("name", container.NewDependencyParam("name"))

	_, err := c.Get("person")
	fmt.Println(err)

	// Output: get("person"): field value "name": getParam("name"): circular dependencies: %name% -> %name%
}

func ExampleContainer_CircularDeps() {
	type Spouse struct {
		Name   string
		Spouse *Spouse
	}

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
	c.OverrideParam("name", container.NewDependencyParam("name"))

	fmt.Println(c.CircularDeps())

	// Output:
	// CircularDeps(): @husband -> @wife -> @husband
	// CircularDeps(): %name% -> %name%
}

func ExampleContainer_Get_setter() {
	mary := container.NewService()
	mary.SetConstructor(func() *Person { // we have to use a pointer, because we use a setter
		return &Person{}
	})
	mary.AppendCall("SetName", container.NewDependencyValue("Mary Jane"))

	c := container.New()
	c.OverrideService("mary", mary)

	maryObject, _ := c.Get("mary")
	fmt.Println(maryObject)

	// Output: &{Mary Jane}
}

func ExampleContainer_Get_wither() {
	mary := container.NewService()
	mary.SetConstructor(func() Person {
		return Person{}
	})
	mary.AppendWither("WithName", container.NewDependencyValue("Mary Jane"))

	c := container.New()
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

func ExampleContainer_AddDecorator() {
	g := container.NewService()
	g.SetValue(greeter{})
	g.Tag("greeter-tag", 0)

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

	num := container.NewService()
	num.SetConstructor(func() int {
		i++
		return i
	})
	num.SetScopeShared()

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

	num := container.NewService()
	num.SetConstructor(func() int {
		i++
		return i
	})
	num.SetScopeNonShared()

	c := container.New()
	c.OverrideService("number", num)

	first, _ := c.Get("number")
	second, _ := c.Get("number")

	// first is not equal to second, because the scope is private
	fmt.Println(first, second)

	// Output: 1 2
}

func ExampleContainer_Get_taggedServices() {
	type Superhero struct {
		Name string
	}

	type Team struct {
		Superheroes []Superhero
	}

	// describe Iron Man
	ironMan := container.NewService()
	ironMan.SetValue(Superhero{
		Name: "Iron Man",
	})
	ironMan.Tag("avengers", 0)

	// describe Thor
	thor := container.NewService()
	thor.SetValue(Superhero{
		Name: "Thor",
	})
	thor.Tag("avengers", 1) // Thor has a higher priority

	// describe Hulk
	hulk := container.NewService()
	hulk.SetValue(Superhero{
		Name: "Hulk",
	})
	hulk.Tag("avengers", 0)

	// describe Avengers
	team := container.NewService()
	team.SetValue(Team{})
	team.SetField("Superheroes", container.NewDependencyTag("avengers"))

	c := container.New()
	c.OverrideService("ironMan", ironMan)
	c.OverrideService("thor", thor)
	c.OverrideService("hulk", hulk)
	c.OverrideService("avengers", team)

	avengers, _ := c.Get("avengers")
	fmt.Printf("%+v\n", avengers)
	// Output: {Superheroes:[{Name:Thor} {Name:Hulk} {Name:Iron Man}]}
}

func ExampleContainer_GetTaggedBy() {
	type Person struct {
		Name string
	}

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

	c := container.New()
	c.OverrideService("p1", p1)
	c.OverrideService("p2", p2)
	c.OverrideService("p3", p3)

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

	s := container.NewService()
	// invalid arguments
	s.SetConstructor(
		func(host string, port int) *Server {
			return &Server{
				Host: host,
				Port: port,
			}
		},
		container.NewDependencyValue(nil),         // it should be a string!
		container.NewDependencyValue("localhost"), // it should be an int!
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

	pi := container.NewService()
	pi.SetValue(math.Pi)
	c.OverrideService("pi", pi)

	three := container.NewService()
	three.SetValue(3)
	three.Tag("int", 0)
	c.OverrideService("three", three)

	fmt.Printf("pi is tagged by int: %v\n", c.IsTaggedBy("pi", "int"))
	fmt.Printf("three is tagged by int: %v\n", c.IsTaggedBy("three", "int"))

	// Output:
	// pi is tagged by int: false
	// three is tagged by int: true
}

func ExampleNewDependencyProvider() {
	c := container.New()
	c.OverrideParam("pi", container.NewDependencyProvider(func() float64 {
		return math.Pi
	}))
	pi, _ := c.GetParam("pi")
	fmt.Printf("%0.2f\n", pi)
	// Output: 3.14
}
