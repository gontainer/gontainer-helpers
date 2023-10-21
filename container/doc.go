/*
Package container provides a set of tools to create a DI Container.

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
*/
package container
