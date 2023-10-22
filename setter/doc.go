/*
Package setter allows for manipulation of a value of an exported field of any struct.

	person := struct {
		name string
	}{}
	err := setter.Set(&person, "name", "Mary", false)
	fmt.Println(person.name)
	fmt.Println(err)
	// Output:
	// Mary
	// <nil>
*/
package setter
