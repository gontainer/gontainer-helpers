package exporter

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	defaultExporter     = newDefaultExporter()
	defaultStringCaster = newChainExporter(
		&boolExporter{},
		&nilExporter{},
		&numberExporter{explicitType: false},
	)
)

func newDefaultExporter() exporter {
	interfaceSliceExporter := &interfaceSliceExporter{}
	primitiveTypeSliceExporter := &primitiveTypeSliceExporter{}

	result := newChainExporter(
		&boolExporter{},
		&nilExporter{},
		&numberExporter{explicitType: true},
		&stringExporter{},
		&bytesExporter{},
		interfaceSliceExporter,
		primitiveTypeSliceExporter,
	)
	interfaceSliceExporter.exporter = result
	primitiveTypeSliceExporter.exporter = result

	return result
}

// Export exports input value to a GO code.
func Export(i any) (string, error) {
	return defaultExporter.export(i)
}

// MustExport exports input value to a GO code.
func MustExport(i any) string {
	r, err := Export(i)
	if err != nil {
		panic(fmt.Sprintf("cannot export %T to string: %s", i, err.Error()))
	}
	return r
}

// CastToString casts input value to a string. This function supports booleans, strings, numeric values and nil-values:
//   - any numeric input returns string that represents its value without a type
//   - any boolean input returns accordingly a string "true" or "false"
//   - any string input results in the output that equals the input
//   - any nil input returns a "nil" string.
func CastToString(i any) (string, error) {
	if r, ok := i.(string); ok {
		return r, nil
	}

	return defaultStringCaster.export(i)
}

// MustCastToString casts input value to a string.
// See CastToString.
func MustCastToString(i any) string {
	r, err := CastToString(i)
	if err != nil {
		panic(fmt.Sprintf("cannot cast %T to string: %s", i, err.Error()))
	}
	return r
}

type exporter interface {
	export(any) (string, error)
}

type subExporter interface {
	exporter
	supports(any) bool
}

type chainExporter struct {
	exporters []subExporter
}

func (c chainExporter) export(v any) (string, error) {
	for _, e := range c.exporters {
		if e.supports(v) {
			return e.export(v)
		}
	}

	return "", fmt.Errorf("type %T is not supported", v)
}

func newChainExporter(exporters ...subExporter) *chainExporter {
	return &chainExporter{exporters: exporters}
}

type boolExporter struct{}

func (boolExporter) export(v any) (string, error) {
	if v == true {
		return "true", nil
	}

	return "false", nil
}

func (boolExporter) supports(v any) bool {
	_, ok := v.(bool)
	return ok
}

type nilExporter struct{}

func (nilExporter) export(any) (string, error) {
	return "nil", nil
}

func (nilExporter) supports(v any) bool {
	return v == nil
}

type numberExporter struct {
	explicitType bool
}

func (n numberExporter) export(v any) (string, error) {
	t := reflect.TypeOf(v)
	var sv string
	switch t.Kind() {
	case reflect.Float32:
		sv = strconv.FormatFloat(float64(v.(float32)), 'f', -1, 32)
	case reflect.Float64:
		sv = strconv.FormatFloat(v.(float64), 'f', -1, 64)
	default:
		sv = fmt.Sprintf("%d", v)
	}
	if n.explicitType {
		sv = fmt.Sprintf("%s(%s)", t.Kind().String(), sv)
	}
	return sv, nil
}

func (n numberExporter) supports(v any) bool {
	t := reflect.TypeOf(v)
	if t == nil {
		return false
	}

	if t.PkgPath() != "" {
		return false
	}

	switch t.Kind() {
	case
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64:
		return true
	}

	return false
}

type stringExporter struct{}

func (stringExporter) export(v any) (string, error) {
	return fmt.Sprintf("%+q", v), nil
}

func (stringExporter) supports(v any) bool {
	_, ok := v.(string)
	return ok
}

type bytesExporter struct{}

func (bytesExporter) export(v any) (string, error) {
	s, _ := stringExporter{}.export(v)
	return fmt.Sprintf("[]byte(%s)", s), nil
}

func (bytesExporter) supports(v any) bool {
	_, ok := v.([]byte)
	return ok
}

type interfaceSliceExporter struct {
	exporter exporter
}

func (i interfaceSliceExporter) export(v any) (string, error) {
	val := reflect.ValueOf(v)
	if val.Type().Kind() == reflect.Slice && val.Len() == 0 {
		return "make([]any, 0)", nil
	}
	parts := make([]string, val.Len())
	for j := 0; j < val.Len(); j++ {
		part, err := i.exporter.export(val.Index(j).Interface())
		if err != nil {
			return "", fmt.Errorf("cannot export %s[%d]: %w", val.Type().Kind().String(), j, err)
		}
		parts[j] = part
	}

	prefix := "[]any"
	if val.Type().Kind() == reflect.Array {
		prefix = fmt.Sprintf("[%d]any", val.Len())
	}

	return prefix + "{" + strings.Join(parts, ", ") + "}", nil
}

func (i interfaceSliceExporter) supports(v any) bool {
	t := reflect.TypeOf(v)
	if t == nil {
		return false
	}
	return t.PkgPath() == "" &&
		(t.Kind() == reflect.Slice || t.Kind() == reflect.Array) &&
		t.Elem().Kind() == reflect.Interface &&
		t.Elem().NumMethod() == 0
}

type primitiveTypeSliceExporter struct {
	exporter exporter
}

func (p primitiveTypeSliceExporter) export(v any) (string, error) {
	val := reflect.ValueOf(v)
	if val.Type().Kind() == reflect.Slice && val.Len() == 0 {
		return fmt.Sprintf("make([]%s, 0)", val.Type().Elem().Kind().String()), nil
	}
	parts := make([]string, val.Len())
	for i := 0; i < val.Len(); i++ {
		var err error
		parts[i], err = p.exporter.export(val.Index(i).Interface())
		if err != nil {
			return "", fmt.Errorf("unexpected err %s[%d]: %w", val.Type().Kind().String(), i, err)
		}
	}
	prefix := "[]"
	if val.Type().Kind() == reflect.Array {
		prefix = fmt.Sprintf("[%d]", val.Len())
	}
	return prefix + val.Type().Elem().Kind().String() + "{" + strings.Join(parts, ", ") + "}", nil
}

func (p primitiveTypeSliceExporter) supports(v any) bool {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Invalid {
		return false
	}
	if val.Type().Kind() != reflect.Slice && val.Type().Kind() != reflect.Array {
		return false
	}
	if val.Type().Elem().PkgPath() != "" {
		return false
	}

	switch val.Type().Elem().Kind() {
	case
		reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64,
		reflect.String:
		return true
	}

	return false
}
