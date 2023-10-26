package reflect

import (
	"fmt"
	"reflect"
)

func convertSlice(from reflect.Value, to reflect.Type) (_ reflect.Value, supports bool, _ error) {
	if isConvertibleSliceOrArray(from, to) {
		v, err := convertSliceOrArray(from, to)
		return v, true, err
	}
	return reflect.Value{}, false, nil
}

func isArrayOrSlice(k reflect.Kind) bool {
	switch k {
	case
		reflect.Slice,
		reflect.Array:
		return true
	}

	return false
}

func isConvertibleSliceOrArray(from reflect.Value, to reflect.Type) bool {
	if !isArrayOrSlice(from.Kind()) || !isArrayOrSlice(to.Kind()) {
		return false
	}

	return true
}

func convertSliceOrArray(from reflect.Value, to reflect.Type) (reflect.Value, error) {
	// check whether slice values are convertible for len == 0
	if from.Len() == 0 && !isAny(from.Type().Elem()) && !isAny(to.Elem()) {
		if _, err := convert(
			reflect.Zero(from.Type().Elem()).Interface(),
			to.Elem(),
		); err != nil {
			return reflect.Value{}, err
		}
	}

	if from.Kind() == reflect.Slice && from.IsNil() {
		// zero value for slice == nil
		return reflect.Zero(to), nil
	}

	var (
		cp    reflect.Value
		toLen int
	)
	if to.Kind() == reflect.Array {
		toLen = to.Len()
		if from.Len() < toLen {
			toLen = from.Len()
		}
		cp = reflect.New(reflect.ArrayOf(to.Len(), to.Elem())).Elem()
	} else {
		toLen = from.Len()
		cp = reflect.MakeSlice(to, from.Len(), from.Len())
	}

	for i := 0; i < toLen; i++ {
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
