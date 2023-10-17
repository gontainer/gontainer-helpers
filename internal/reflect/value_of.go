package reflect

import (
	"fmt"
	"reflect"
)

// ValueOf is an extended version of [reflect.ValueOf].
//
// Built-in [reflect.ValueOf](nil) returns the zero [reflect.Value].
// ValueOf for the `i == nil` and a nil-able [reflect.Kind] of `t` returns a zero value from `t`.
func ValueOf(i interface{}, t reflect.Type, convert_ bool) (reflect.Value, error) {
	if convert_ {
		return convert(i, t)
	}

	r := reflect.ValueOf(i)
	if !r.IsValid() {
		return zeroForNilable(i, t)
	}

	return r, nil
}

func zeroForNilable(i interface{}, t reflect.Type) (reflect.Value, error) {
	if i == nil && isNilable(t.Kind()) {
		return reflect.Zero(t), nil
	}

	return reflect.Value{}, fmt.Errorf("cannot cast `%T` to `%s`", i, t.String())
}

func isNilable(k reflect.Kind) bool {
	switch k {
	case
		reflect.Chan,
		reflect.Func,
		reflect.Map,
		reflect.Ptr,
		reflect.UnsafePointer,
		reflect.Interface,
		reflect.Slice:
		return true
	}
	return false
}
