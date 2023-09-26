package exporter

import (
	"fmt"
	"reflect"
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

// Export exports input value to a GO code.
func Export(i interface{}) (string, error) {
	return defaultExporter.export(i)
}

// MustExport exports input value to a GO code.
func MustExport(i interface{}) string {
	r, err := Export(i)
	if err != nil {
		panic(fmt.Sprintf("cannot export `%T` to string: %s", i, err.Error()))
	}
	return r
}

// ToString casts input value to a string. This function supports booleans, strings, numeric values and nil-values:
//   - any numeric input returns string that represents its value without a type
//   - any boolean input returns accordingly a string "true" or "false"
//   - any string input results in the output that equals the input
//   - any nil input returns a "nil" string.
func ToString(i interface{}) (string, error) {
	if r, ok := i.(string); ok {
		return r, nil
	}

	return defaultStringCaster.export(i)
}

// MustToString casts input value to a string.
// See ToString.
func MustToString(i interface{}) string {
	r, err := ToString(i)
	if err != nil {
		panic(fmt.Sprintf("cannot cast `%T` to string: %s", i, err.Error()))
	}
	return r
}

type exporter interface {
	export(interface{}) (string, error)
}

type subExporter interface {
	exporter
	supports(interface{}) bool
}

type chainExporter struct {
	exporters []subExporter
}

func (c chainExporter) export(v interface{}) (string, error) {
	for _, e := range c.exporters {
		if e.supports(v) {
			return e.export(v)
		}
	}

	return "", fmt.Errorf("type `%T` is not supported", v)
}

func newDefaultExporter() exporter {
	interfaceSliceExporter := newInterfaceSliceExporter(nil)
	primitiveTypeSliceExporter := newPrimitiveTypeSliceExporter(nil)

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

func newChainExporter(exporters ...subExporter) *chainExporter {
	return &chainExporter{exporters: exporters}
}

type boolExporter struct{}

func (b boolExporter) export(v interface{}) (string, error) {
	if v == true {
		return "true", nil
	}

	return "false", nil
}

func (b boolExporter) supports(v interface{}) bool {
	_, ok := v.(bool)
	return ok
}

type nilExporter struct{}

func (n nilExporter) export(interface{}) (string, error) {
	return "nil", nil
}

func (n nilExporter) supports(v interface{}) bool {
	return v == nil
}

type numberExporter struct {
	explicitType bool
}

func (n numberExporter) export(v interface{}) (string, error) {
	t := reflect.TypeOf(v)
	var sv string
	switch t.Kind() {
	case
		reflect.Float32,
		reflect.Float64:
		sv = fmt.Sprintf("%#v", v)
	default:
		sv = fmt.Sprintf("%d", v)
	}
	if n.explicitType {
		sv = fmt.Sprintf("%s(%s)", t.Kind().String(), sv)
	}
	return sv, nil
}

func (n numberExporter) supports(v interface{}) bool {
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

func (s stringExporter) export(v interface{}) (string, error) {
	return fmt.Sprintf("%+q", v), nil
}

func (s stringExporter) supports(v interface{}) bool {
	_, ok := v.(string)
	return ok
}

type bytesExporter struct{}

func (b bytesExporter) export(v interface{}) (string, error) {
	s, _ := stringExporter{}.export(v)
	return fmt.Sprintf("[]byte(%s)", s), nil
}

func (b bytesExporter) supports(v interface{}) bool {
	_, ok := v.([]byte)
	return ok
}

type interfaceSliceExporter struct {
	exporter exporter
}

func newInterfaceSliceExporter(exporter exporter) *interfaceSliceExporter {
	return &interfaceSliceExporter{exporter: exporter}
}

func (i interfaceSliceExporter) export(v interface{}) (string, error) {
	val := reflect.ValueOf(v)
	if val.Type().Kind() == reflect.Slice && val.Len() == 0 {
		return "make([]interface{}, 0)", nil
	}
	parts := make([]string, val.Len())
	for j := 0; j < val.Len(); j++ {
		part, err := i.exporter.export(val.Index(j).Interface())
		if err != nil {
			return "", fmt.Errorf("cannot export %s[%d]: %w", val.Type().Kind().String(), j, err)
		}
		parts[j] = part
	}

	prefix := "[]interface{}"
	if val.Type().Kind() == reflect.Array {
		prefix = fmt.Sprintf("[%d]interface{}", val.Len())
	}

	return prefix + "{" + strings.Join(parts, ", ") + "}", nil
}

func (i interfaceSliceExporter) supports(v interface{}) bool {
	t := reflect.TypeOf(v)
	if t == nil {
		return false
	}
	return t.PkgPath() == "" &&
		(t.Kind() == reflect.Slice || t.Kind() == reflect.Array) &&
		t.Elem().Kind() == reflect.Interface
}

type primitiveTypeSliceExporter struct {
	exporter exporter
}

func newPrimitiveTypeSliceExporter(exporter exporter) *primitiveTypeSliceExporter {
	return &primitiveTypeSliceExporter{exporter: exporter}
}

func (p primitiveTypeSliceExporter) export(v interface{}) (string, error) {
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

func (p primitiveTypeSliceExporter) supports(v interface{}) bool {
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
