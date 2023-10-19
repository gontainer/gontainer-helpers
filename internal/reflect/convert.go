package reflect

import (
	"fmt"
	"reflect"
)

// convert converts given value to given type whenever it is possible.
// In opposition to built-in reflect package it allows to convert []any to []type and []type to []any.
func convert(value any, to reflect.Type) (reflect.Value, error) {
	from := reflect.ValueOf(value)
	if !from.IsValid() {
		return zeroForNilable(value, to)
	}

	if to.Kind() == reflect.Array &&
		from.Kind() == reflect.Slice &&
		from.Len() > to.Len() {
		return reflect.Value{}, fmt.Errorf("cannot cast `%T` (length %d) to `%s`", value, from.Len(), to.String())
	}

	if from.Type().ConvertibleTo(to) {
		return from.Convert(to), nil
	}

	if !isConvertibleSliceOrArray(from.Type(), to) {
		return reflect.Value{}, fmt.Errorf("cannot cast `%s` to `%s`", from.Type().String(), to.String())
	}

	slice, err := convertSliceOrArray(from, to)
	if err != nil {
		return reflect.Value{}, fmt.Errorf("cannot cast `%s` to `%s`: %w", from.Type().String(), to.String(), err)
	}

	return slice, nil
}

func isConvertibleSliceOrArray(from reflect.Type, to reflect.Type) bool {
	if (from.Kind() != reflect.Slice && from.Kind() != reflect.Array) ||
		(to.Kind() != reflect.Slice && to.Kind() != reflect.Array) {
		return false
	}

	if from.Kind() == reflect.Array && to.Kind() == reflect.Array && from.Len() > to.Len() {
		return false
	}

	if from.Elem().Kind() == reflect.Interface || to.Elem().Kind() == reflect.Interface {
		return true
	}

	if from.Elem().ConvertibleTo(to.Elem()) {
		return true
	}

	if isConvertibleSliceOrArray(from.Elem(), to.Elem()) {
		return true
	}

	return false
}

func convertSliceOrArray(from reflect.Value, to reflect.Type) (reflect.Value, error) {
	var cp reflect.Value
	if to.Kind() == reflect.Array {
		cp = reflect.New(reflect.ArrayOf(to.Len(), to.Elem())).Elem()
	} else {
		cp = reflect.MakeSlice(to, from.Len(), from.Cap())
	}

	for i := 0; i < from.Len(); i++ {
		item := from.Index(i)
		for item.Kind() == reflect.Interface {
			item = item.Elem()
		}
		var currVal any = nil
		if item.IsValid() {
			currVal = item.Interface()
		}
		curr, err := convert(currVal, to.Elem())
		if err != nil {
			return reflect.Value{}, fmt.Errorf("%d: %w", i, err)
		}
		cp.Index(i).Set(curr)
	}
	return cp, nil
}
