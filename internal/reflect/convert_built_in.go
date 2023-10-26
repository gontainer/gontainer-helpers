package reflect

import (
	"reflect"
)

func convertBuiltIn(from reflect.Value, to reflect.Type) (_ reflect.Value, supports bool, _ error) {
	// avoid a panic, see [reflect.Type.ConvertibleTo]
	if to.Kind() == reflect.Array && (from.Kind() == reflect.Slice || from.Kind() == reflect.Array) && from.Len() < to.Len() {
		return reflect.Value{}, false, nil
	}

	if from.Type().ConvertibleTo(to) {
		return from.Convert(to), true, nil
	}

	return reflect.Value{}, false, nil
}

func isAny(v reflect.Type) bool {
	if v.Kind() != reflect.Interface {
		return false
	}

	if v.NumMethod() > 0 {
		return false
	}

	return true
}
