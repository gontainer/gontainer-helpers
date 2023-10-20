package setter_test

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/setter"
)

func ExampleSet_ok() {
	person := struct {
		name string
	}{}
	err := setter.Set(&person, "name", "Mary")
	fmt.Println(person.name)
	fmt.Println(err)
	// Output:
	// Mary
	// <nil>
}

func ExampleSet_errFieldDoesNotExists() {
	person := struct {
		name string
	}{}
	err := setter.Set(&person, "firstname", "Mary")
	fmt.Println(err)
	// Output: set (*struct { name string })."firstname": field "firstname" does not exist
}

func ExampleSet_errNoPtr() {
	person := struct {
		name string
	}{}
	err := setter.Set(person, "name", "Mary")
	fmt.Println(err)
	// Output: expected pointer to struct, struct given
}
