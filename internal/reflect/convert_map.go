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

	// TODO check whether keys are convertible, if no
	// supports = false; return

	mapType := reflect.MapOf(to.Key(), to.Elem())
	result := reflect.MakeMap(mapType)

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
