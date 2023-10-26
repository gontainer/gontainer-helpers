package reflect

import (
	"fmt"
	"reflect"
)

func convertMap(from reflect.Value, to reflect.Type) (_ reflect.Value, supports bool, _ error) {
	if from.Type().Kind() != reflect.Map || to.Kind() != reflect.Map {
		supports = false
		return
	}

	if from.Len() == 0 {
		if !isAny(from.Type().Key()) || !isAny(to.Key()) {
			if _, err := convert(
				reflect.Zero(from.Type().Key()).Interface(),
				to.Key(),
			); err != nil {
				return reflect.Value{}, true, fmt.Errorf("non convertible keys: %w", err)
			}
		}
		if !isAny(from.Type().Elem()) || !isAny(to.Elem()) {
			if _, err := convert(
				reflect.Zero(from.Type().Elem()).Interface(),
				to.Elem(),
			); err != nil {
				return reflect.Value{}, true, fmt.Errorf("non convertible values: %w", err)
			}
		}
	}

	if from.IsNil() {
		return reflect.Zero(to), true, nil
	}

	mapType := reflect.MapOf(to.Key(), to.Elem())
	result := reflect.MakeMapWithSize(mapType, from.Len())

	iter := from.MapRange()
	for iter.Next() {
		newKey, err := convert(iter.Key().Interface(), to.Key())
		if err != nil {
			err = fmt.Errorf("map key: %w", err)
			return reflect.Value{}, true, err
		}

		newValue, err := convert(iter.Value().Interface(), to.Elem())
		if err != nil {
			err = fmt.Errorf("map value: %w", err)
			return reflect.Value{}, true, err
		}

		result.SetMapIndex(newKey, newValue)
	}

	return result, true, nil
}
