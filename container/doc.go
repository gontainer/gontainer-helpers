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
