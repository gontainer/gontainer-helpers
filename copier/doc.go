/*
Package copier allows for copying a value to a variable with an unknown type. It can also convert a value to another type.

	var (
		from = 5 // the type of the variable `to` can be different from the type of the variable `from`
		to   any // as long as the value of the `from` is assignable to the `to`
	)
	_ = copier.Copy(from, &to, false)
	fmt.Println(to)
	// Output: 5
*/
package copier
