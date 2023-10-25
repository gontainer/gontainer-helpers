package copier_test

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/v2/copier"
)

func ExampleCopy_convertOK() {
	var (
		from = int(5) // uint is not assignable to int,
		to   uint     // but [copier.Copy] can convert the type
	)
	err := copier.Copy(from, &to, true)
	fmt.Println(to)
	fmt.Println(err)
	// Output:
	// 5
	// <nil>
}

func ExampleCopy_convertMap() {
	var (
		from = map[int64]any{0: "Jane", 1: "John"}
		to   map[int32]string // let's convert keys and values
	)
	err := copier.Copy(from, &to, true)
	fmt.Println(to)
	fmt.Println(err)
	// Output:
	// map[0:Jane 1:John]
	// <nil>
}

func ExampleCopy_ok() {
	var (
		from = 5 // the type of the variable `to` can be different from the type of the variable `from`
		to   any // as long as the value of the `from` is assignable to the `to`
	)
	err := copier.Copy(from, &to, false)
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
	err := copier.Copy(from, &to, false)
	fmt.Println(to)
	fmt.Println(err)
	// Output:
	// 0
	// value of type int is not assignable to type uint
}

func ExampleCopy_err2() {
	var (
		from float32 = 5
		to   uint    = 0
	)
	err := copier.Copy(from, &to, false)
	fmt.Println(to)
	fmt.Println(err)
	// Output:
	// 0
	// value of type float32 is not assignable to type uint
}

func ExampleCopy_err3() {
	var (
		from *int
		to   *uint
	)
	err := copier.Copy(from, &to, false)
	fmt.Println(to)
	fmt.Println(err)
	// Output:
	// <nil>
	// value of type *int is not assignable to type *uint
}
