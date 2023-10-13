package copier_test

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/copier"
)

func ExampleCopy_ok() {
	var (
		from = 5         // the type of the variable `to` can be different from the type of the variable `from`
		to   interface{} // as long as the value of `from` is assignable to the `to`
	)
	err := copier.Copy(from, &to)
	fmt.Println(to)
	fmt.Println(err)
	// Output:
	// 5
	// <nil>
}

func ExampleCopy_err() {
	var (
		from float32 = 5
		to   uint    = 0
	)
	err := copier.Copy(from, &to)
	fmt.Println(to)
	fmt.Println(err)
	// Output:
	// 0
	// reflect.Set: value of type float32 is not assignable to type uint
}
