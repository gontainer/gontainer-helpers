package exporter_test

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/exporter"
)

func ExampleCastToString_string() {
	s, _ := exporter.CastToString("hello world")
	fmt.Println(s)
	// Output: hello world
}

func ExampleCastToString_bool() {
	s, _ := exporter.CastToString(false)
	fmt.Println(s)
	// Output: false
}

func ExampleCastToString_pi() {
	s, _ := exporter.CastToString(float32(3.1416))
	fmt.Println(s)
	// Output: 3.1416
}

func ExampleCastToString_nil() {
	s, _ := exporter.CastToString(nil)
	fmt.Println(s)
	// Output: nil
}

func ExampleCastToString_notSupported() {
	_, err := exporter.CastToString(struct{}{})
	fmt.Println(err)
	// Output: type `struct {}` is not supported
}

func ExampleExport_int() {
	s, _ := exporter.Export(5)
	fmt.Println(s)
	// Output: int(5)
}

func ExampleExport_pi() {
	s, _ := exporter.Export(float32(3.1416))
	fmt.Println(s)
	// Output: float32(3.1416)
}

func ExampleExport_string() {
	s, _ := exporter.Export("hello world")
	fmt.Println(s)
	// Output: "hello world"
}

func ExampleExport_slice() {
	s, _ := exporter.Export([]uint{1, 2, 3})
	fmt.Println(s)
	// Output: []uint{uint(1), uint(2), uint(3)}
}

func ExampleExport_slice2() {
	s, _ := exporter.Export([]any{int32(1), int64(2), float32(3.14), "hello world"})
	fmt.Println(s)
	// Output: []any{int32(1), int64(2), float32(3.14), "hello world"}
}

func ExampleExport_emptySlice() {
	var v any = make([]any, 0)
	s, _ := exporter.Export(v)
	fmt.Println(s)
	// Output: make([]any, 0)
}

func ExampleExport_emptySlice2() {
	var v []any = nil
	s, _ := exporter.Export(v)
	fmt.Println(s)
	// Output: make([]any, 0)
}

func ExampleExport_array() {
	s, _ := exporter.Export([3]uint{1, 2, 3})
	fmt.Println(s)
	// Output: [3]uint{uint(1), uint(2), uint(3)}
}

func ExampleExport_array2() {
	s, _ := exporter.Export([3]any{nil, 1.5, "hello world"})
	fmt.Println(s)
	// Output: [3]any{nil, float64(1.5), "hello world"}
}

func ExampleExport_emptyArray() {
	s, _ := exporter.Export([0]uint{})
	fmt.Println(s)
	// Output: [0]uint{}
}

func ExampleExport_emptyArray2() {
	s, _ := exporter.Export([0]any{})
	fmt.Println(s)
	// Output: [0]any{}
}

func ExampleExport_err() {
	_, err := exporter.Export(struct{}{})
	fmt.Println(err)
	// Output: type `struct {}` is not supported
}
