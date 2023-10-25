package reflect

import (
	"reflect"
)

func convertBuiltIn(from reflect.Value, to reflect.Type) (_ reflect.Value, supports bool, _ error) {
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
