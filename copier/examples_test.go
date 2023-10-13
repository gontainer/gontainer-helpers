package copier_test

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/copier"
)

func ExampleConvertAndCopy_ok() {
	var (
		from = int(5) // uint is not assignable to int,
		to   uint     // but ConvertAndCopy can convert the type
	)
	err := copier.ConvertAndCopy(from, &to)
	fmt.Println(to)
	fmt.Println(err)
	// Output:
	// 5
	// <nil>
}

func ExampleCopy_ok() {
	var (
		from = 5         // the type of the variable `to` can be different from the type of the variable `from`
		to   interface{} // as long as the value of the `from` is assignable to the `to`
	)
	err := copier.Copy(from, &to)
	fmt.Println(to)
	fmt.Println(err)
	// Output:
	// 5
	// <nil>
}

func ExampleCopy_err1() {
	var (
		from = int(5)
		to   uint
	)
	err := copier.Copy(from, &to)
	fmt.Println(to)
	fmt.Println(err)
	// Output:
	// 0
	// reflect.Set: value of type int is not assignable to type uint
}

func ExampleCopy_err2() {
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
