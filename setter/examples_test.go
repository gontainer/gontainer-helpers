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
	type Person struct {
		name string //nolint:unused
	}
	var person Person
	err := setter.Set(person, "name", "Mary", false)
	fmt.Println(err)
	// Output: set (setter_test.Person)."name": expected pointer to struct, struct given
}

func ExampleSet_convert() {
	person := struct {
		age int32
	}{}
	_ = setter.Set(&person, "age", uint8(30), true)
	fmt.Printf("%#v\n", person)
	// Output: struct { age int32 }{age:30}
}
