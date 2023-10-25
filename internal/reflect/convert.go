package reflect

import (
	"fmt"
	"reflect"
)

type converter interface {
	convert(from reflect.Value, to reflect.Type) (_ reflect.Value, supports bool, _ error)
}

type converterFn func(from reflect.Value, to reflect.Type) (_ reflect.Value, supports bool, _ error)

func (fn converterFn) convert(from reflect.Value, to reflect.Type) (_ reflect.Value, supports bool, _ error) {
	return fn(from, to)
}

var (
	converters []converter
)

func init() {
	converters = []converter{
		converterFn(convertBuiltIn),
		converterFn(convertSlice),
		converterFn(convertMap),
	}
}

// convert converts given value to given type whenever it is possible.
// In opposition to built-in reflect package it can convert slices and maps.
func convert(value any, to reflect.Type) (reflect.Value, error) {
	from := reflect.ValueOf(value)
	if !from.IsValid() {
		return zeroForNilable(value, to)
	}

	for _, c := range converters {
		if v, supports, err := c.convert(from, to); supports {
			if err != nil {
				err = fmt.Errorf("cannot convert %s to %s: %w", from.Type().String(), to.String(), err)
			}
			return v, err
		}
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %s to %s", from.Type().String(), to.String())
}
