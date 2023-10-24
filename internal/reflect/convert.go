package reflect

import (
	"fmt"
	"reflect"
)

type preChecker interface {
	check(from reflect.Value, to reflect.Type) error
}

type preCheckFn func(from reflect.Value, to reflect.Type) error

func (fn preCheckFn) check(from reflect.Value, to reflect.Type) error {
	return fn(from, to)
}

type converter interface {
	convert(from reflect.Value, to reflect.Type) (_ reflect.Value, supports bool, _ error)
}

type converterFn func(from reflect.Value, to reflect.Type) (_ reflect.Value, supports bool, _ error)

func (fn converterFn) convert(from reflect.Value, to reflect.Type) (_ reflect.Value, supports bool, _ error) {
	return fn(from, to)
}

var (
	preCheckers []preChecker
	converters  []converter
)

func init() {
	preCheckers = []preChecker{
		preCheckFn(func(from reflect.Value, to reflect.Type) error {
			if to.Kind() == reflect.Array &&
				from.Kind() == reflect.Slice &&
				from.Len() > to.Len() {
				return fmt.Errorf("cannot convert %T (length %d) to %s", from.Interface(), from.Len(), to.String())
			}
			return nil
		}),
	}

	converters = []converter{
		converterFn(convertBuiltIn),
		converterFn(convertSlice),
		converterFn(convertMap),
	}
}

func convertBuiltIn(from reflect.Value, to reflect.Type) (_ reflect.Value, supports bool, _ error) {
	if from.Type().ConvertibleTo(to) {
		return from.Convert(to), true, nil
	}

	return reflect.Value{}, false, nil
}

// convert converts given value to given type whenever it is possible.
// In opposition to built-in reflect package it allows to convert []any to []type and []type to []any.
func convert(value any, to reflect.Type) (reflect.Value, error) {
	from := reflect.ValueOf(value)
	if !from.IsValid() {
		return zeroForNilable(value, to)
	}

	for _, pc := range preCheckers {
		if err := pc.check(from, to); err != nil {
			return reflect.Value{}, err
		}
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
