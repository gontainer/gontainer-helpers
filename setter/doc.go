/*
Package setter allows for manipulation of a value of a field of any struct.
The third argument instructs the setter whether you allow for converting the type.

	person := struct {
		name string
	}{}
	_ = setter.Set(&person, "name", "Mary", false)
	fmt.Println(person.name)
	// Output: Mary
*/
package setter
