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

package accessors_test

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/v2/accessors"
)

func Example() {
	person := struct {
		name string
	}{}
	_ = accessors.Set(&person, "name", "Mary", false)
	fmt.Println(person.name)
	// Output: Mary
}

func ExampleGet() {
	person := struct {
		name string
	}{
		name: "Mary",
	}
	v, _ := accessors.Get(person, "name")
	fmt.Println(v)
	// Output: Mary
}

func ExampleSet_ok() {
	person := struct {
		name string
	}{}
	err := accessors.Set(&person, "name", "Mary", false)
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
	err := accessors.Set(&person, "firstname", "Mary", false)
	fmt.Println(err)
	// Output: set (*struct { name string })."firstname": field "firstname" does not exist
}

func ExampleSet_errNoPtr() {
	type Person struct {
		name string //nolint:unused
	}
	var person Person
	err := accessors.Set(person, "name", "Mary", false)
	fmt.Println(err)
	// Output: set (accessors_test.Person)."name": expected pointer to struct, accessors_test.Person given
}

func ExampleSet_typeMismatchingError() {
	type name string

	type Person struct {
		name string //nolint:unused
	}
	var person Person
	err := accessors.Set(&person, "name", name("Jane"), false)
	fmt.Println(err)
	// Output: set (*accessors_test.Person)."name": value of type accessors_test.name is not assignable to type string
}

func ExampleSet_typeMismatchingConvert() {
	type name string

	type Person struct {
		name string //nolint:unused
	}
	var person Person
	_ = accessors.Set(&person, "name", name("Jane"), true)
	fmt.Println(person.name)
	// Output: Jane
}
