package setter_test

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/v2/setter"
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
	type Person struct {
		name string //nolint:unused
	}
	var person Person
	err := setter.Set(person, "name", "Mary", false)
	fmt.Println(err)
	// Output: set (setter_test.Person)."name": expected pointer to struct, struct given
}

func ExampleSet_typeMismatchingError() {
	type name string

	type Person struct {
		name string //nolint:unused
	}
	var person Person
	err := setter.Set(&person, "name", name("Jane"), false)
	fmt.Println(err)
	// Output: set (*setter_test.Person)."name": value of type setter_test.name is not assignable to type string
}

func ExampleSet_typeMismatchingConvert() {
	type name string

	type Person struct {
		name string //nolint:unused
	}
	var person Person
	_ = setter.Set(&person, "name", name("Jane"), true)
	fmt.Println(person.name)
	// Output: Jane
}
