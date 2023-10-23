/*
Package setter allows for manipulation of a value of an exported field of any struct.

	person := struct {
		name string
	}{}
	_ = setter.Set(&person, "name", "Mary", false)
	fmt.Println(person.name)
	// Output: Mary
*/
package setter
