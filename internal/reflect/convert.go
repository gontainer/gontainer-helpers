package reflect

import (
	"fmt"
	"reflect"
)

// Convert converts given value to given type whenever it is possible.
// In opposition to built-in reflect package it allows to convert []interface{} to []type and []type to []interface{}.
func Convert(value interface{}, to reflect.Type) (reflect.Value, error) {
	// it is required to avoid panic (reflect: call of reflect.Value.Type on zero Value)
	// in case of the following code
	// caller.Call(func(v interface{}) { fmt.Println(v) }, nil)
	if value == nil {
		if IsNilable(to.Kind()) {
			return reflect.Zero(to), nil
		}
		return reflect.Value{}, fmt.Errorf("cannot cast `%T` to `%s`", value, to.String())
	}
	from := reflect.ValueOf(value)
	if from.Type().ConvertibleTo(to) {

		// TODO check all edge cases
		if to.Kind() == reflect.Array &&
			from.Kind() == reflect.Slice &&
			from.Len() > to.Len() {
			return reflect.Value{}, fmt.Errorf("cannot cast `%T` (len %d) to `%s`", value, from.Len(), to.String())
		}

		return from.Convert(to), nil
	}

	if !isConvertibleSlice(from.Type(), to) {
		return reflect.Value{}, fmt.Errorf("cannot cast `%s` to `%s`", from.Type().String(), to.String())
	}

	slice, err := convertSlice(from, to)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("cannot cast `%s` to `%s`: %w", from.Type().String(), to.String(), err)
	}

	return slice, nil
}

func isConvertibleSlice(from reflect.Type, to reflect.Type) bool {
	if from.Kind() != reflect.Slice || to.Kind() != reflect.Slice {
		return false
	}

	if from.Elem().Kind() == reflect.Interface || to.Elem().Kind() == reflect.Interface {
		return true
	}

	if from.Elem().ConvertibleTo(to.Elem()) {
		return true
	}

	if isConvertibleSlice(from.Elem(), to.Elem()) {
		return true
	}

	return false
}

func convertSlice(from reflect.Value, to reflect.Type) (reflect.Value, error) {
	cp := reflect.MakeSlice(to, 0, 0)
	for i := 0; i < from.Len(); i++ {
		item := from.Index(i)
		for item.Kind() == reflect.Interface {
			item = item.Elem()
		}
		var currVal interface{} = nil
		if item.IsValid() {
			currVal = item.Interface()
		}
		curr, err := Convert(currVal, to.Elem())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("%d: %w", i, err)
		}
		cp = reflect.Append(cp, curr)
	}
	return cp, nil
}
