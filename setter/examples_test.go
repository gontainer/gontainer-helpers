package setter_test

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/setter"
)

func ExampleSet_ok() {
	person := struct {
		name string
	}{}
	err := setter.Set(&person, "name", "Mary", false)
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
	err := setter.Set(&person, "firstname", "Mary", false)
	fmt.Println(err)
	// Output: set (*struct { name string })."firstname": field "firstname" does not exist
}

func ExampleSet_errNoPtr() {
	person := struct {
		name string
	}{}
	err := setter.Set(person, "name", "Mary", false)
	fmt.Println(err)
	// Output: expected pointer to struct, struct given
}
