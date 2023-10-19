package caller_test

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/caller"
)

type Person struct {
	name string
}

func NewPerson(name string) *Person {
	return &Person{name: name}
}

func (p *Person) SetName(n string) {
	p.name = n
}

func (p Person) WithName(n string) Person {
	p.name = n
	return p
}

func ExampleCall_ok() {
	// type Person struct {
	//     name string
	// }
	//
	// func (p *Person) SetName(n string) {
	//     p.name = n
	// }

	p := &Person{}
	_, _ = caller.Call(p.SetName, []interface{}{"Mary"}, false)
	fmt.Println(p.name)
	// Output: Mary
}

func ExampleCall_returnValue() {
	fn := func(a int, b int) int {
		return a * b
	}
	r, _ := caller.Call(fn, []interface{}{3, 2}, false)
	fmt.Println(r[0])
	// Output: 6
}

// TODO: panic: reflect: Call using string as type int
//func ExampleCall_error() {
//	fn := func(a int, b int) int {
//		return a * b
//	}
//	_, err := caller.Call(fn, []interface{}{"2", "2"}, false)
//	fmt.Println(err)
//	// Output:
//	// arg0: cannot cast `string` to `int`
//	// arg1: cannot cast `string` to `int`
//}

func ExampleCallProvider() {
	// type Person struct {
	//     Name string
	// }
	//
	// func NewPerson(name string) *Person {
	//     return &Person{Name: name}
	// }

	p, _ := caller.CallProvider(NewPerson, []interface{}{"Mary"}, false)
	fmt.Printf("%+v", p)
	// Output: &{name:Mary}
}

func ExampleCallByName() {
	// type Person struct {
	//     name string
	// }
	//
	// func (p *Person) SetName(n string) {
	//     p.name = n
	// }

	p := &Person{}
	_, _ = caller.CallByName(p, "SetName", []interface{}{"Mary"}, false)
	fmt.Println(p.name)
	// Output: Mary
}

func ExampleCallWitherByName() {
	// type Person struct {
	//     name string
	// }
	//
	// func (p Person) WithName(n string) Person {
	//     p.name = n
	//     return p
	// }

	p := Person{}
	p2, _ := caller.CallWitherByName(p, "WithName", []interface{}{"Mary"}, false)
	fmt.Printf("%+v", p2)
	// Output: {name:Mary}
}
