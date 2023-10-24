package reflect

import (
	"fmt"
	"reflect"
)

func convertSlice(from reflect.Value, to reflect.Type) (_ reflect.Value, supports bool, _ error) {
	if isConvertibleSliceOrArray(from.Type(), to) {
		v, err := convertSliceOrArray(from, to)
		return v, true, err
	}
	return reflect.Value{}, false, nil
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

	if isConvertibleSliceOrArray(from.Elem(), to.Elem()) { // TODO it won't work with maps
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

	// TODO convert nil to nil
	// TODO check whether values are convertible
	if from.Kind() == reflect.Slice && from.IsNil() {
		return reflect.Zero(to), nil
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
			return reflect.Value{}, fmt.Errorf("#%d: %w", i, err)
		}
		cp.Index(i).Set(curr)
	}
	return cp, nil
}
