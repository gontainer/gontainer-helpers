package reflect

import (
	"fmt"
	"reflect"
)

/*
ValueOf is an extended version of [reflect.ValueOf].

Built-in [reflect.ValueOf](nil) returns the zero [reflect.Value].
ValueOf for the `i == nil` and a nil-able [reflect.Kind] of `to` returns a zero value from `to`.

If `result.Type()` is not assignable to `to` it returns an error.
*/
func ValueOf(i any, to reflect.Type, convert_ bool) (result reflect.Value, err error) {
	if convert_ {
		return convert(i, to)
	}

	defer func() {
		if err == nil {
			err = isAssignable(result.Type(), to)
		}
	}()

	r := reflect.ValueOf(i)
	if !r.IsValid() {
		return zeroForNilable(i, to)
	}

	return r, nil
}

func zeroForNilable(i any, t reflect.Type) (reflect.Value, error) {
	if i == nil && isNilable(t.Kind()) {
		return reflect.Zero(t), nil
	}

	// TODO: unify this error with isNilable
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

func isAssignable(from, to reflect.Type) error {
	if !from.AssignableTo(to) {
		return fmt.Errorf("value of type %s is not assignable to type %s", from.String(), to.String())
	}
	return nil
}
