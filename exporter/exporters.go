// Copyright (c) 2023 Bartłomiej Krukowski
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is furnished
// to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package exporter

import (
	"errors"
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
	return newDisposableExporter(func() exporter {
		interfaceSliceExp := &interfaceSliceExporter{}
		primitiveTypeSliceExp := &primitiveTypeSliceExporter{}
		multiArrayExp := &multiArray{}

		result := newAntiLoopExporter(newChainExporter(
			&boolExporter{},
			&nilExporter{},
			&numberExporter{explicitType: true},
			&stringExporter{},
			&bytesExporter{},
			interfaceSliceExp,
			primitiveTypeSliceExp,
			multiArrayExp,
		))

		interfaceSliceExp.exporter = result
		primitiveTypeSliceExp.exporter = result
		multiArrayExp.exporter = result

		return result
	})
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
//   - any nil input returns a "nil" string
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
	supports(any) bool
}

type disposableExporter struct {
	factory func() exporter
}

func newDisposableExporter(factory func() exporter) *disposableExporter {
	return &disposableExporter{factory: factory}
}

func (d disposableExporter) export(a any) (string, error) {
	return d.factory().export(a)
}

func (d disposableExporter) supports(a any) bool {
	return d.factory().supports(a)
}

type stack []any

func newStack() *stack {
	r := make(stack, 0)
	return &r
}

func (s *stack) pop() {
	*s = (*s)[:len(*s)-1]
}

func (s *stack) push(v any) error {
	for _, x := range *s {
		if reflect.DeepEqual(x, v) {
			return errors.New("unexpected infinite loop")
		}
	}
	*s = append(*s, v)
	return nil
}

type antiLoopExporter struct {
	stack *stack
	next  exporter
}

func (a antiLoopExporter) export(v any) (string, error) {
	if err := a.stack.push(v); err != nil {
		return "", err
	}
	defer a.stack.pop()
	return a.next.export(v)
}

func (a antiLoopExporter) supports(v any) bool {
	return a.next.supports(v)
}

func newAntiLoopExporter(next exporter) *antiLoopExporter {
	return &antiLoopExporter{stack: newStack(), next: next}
}

type chainExporter struct {
	exporters []exporter
}

func (c chainExporter) export(v any) (string, error) {
	for _, e := range c.exporters {
		if e.supports(v) {
			return e.export(v)
		}
	}

	return "", fmt.Errorf("type %T is not supported", v)
}

func (c chainExporter) supports(v any) bool {
	for _, e := range c.exporters {
		if e.supports(v) {
			return true
		}
	}
	return false
}

func newChainExporter(exporters ...exporter) *chainExporter {
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
	if val.Type().Kind() == reflect.Slice {
		switch {
		case val.IsNil():
			return "([]interface{})(nil)", nil
		case val.Len() == 0:
			return "make([]interface{}, 0)", nil
		}
	}

	prefix := "[]interface{}"
	if val.Type().Kind() == reflect.Array {
		prefix = fmt.Sprintf("[%d]interface{}", val.Len())
	}

	parts := make([]string, val.Len())
	for j := 0; j < val.Len(); j++ {
		part, err := i.exporter.export(val.Index(j).Interface())
		if err != nil {
			return "", fmt.Errorf("cannot export (%s)[%d]: %w", prefix, j, err)
		}
		parts[j] = part
	}

	return prefix + "{" + strings.Join(parts, ", ") + "}", nil
}

func (i interfaceSliceExporter) supports(v any) bool {
	t := reflect.TypeOf(v)
	if t == nil {
		return false
	}
	return isBuiltInSliceOrArray(t) && isAny(t.Elem())
}

type primitiveTypeSliceExporter struct {
	exporter exporter
}

func (p primitiveTypeSliceExporter) export(v any) (string, error) {
	val := reflect.ValueOf(v)
	if val.Type().Kind() == reflect.Slice {
		switch {
		case val.IsNil():
			return fmt.Sprintf("([]%s)(nil)", val.Type().Elem().Kind().String()), nil
		case val.Len() == 0:
			return fmt.Sprintf("make([]%s, 0)", val.Type().Elem().Kind().String()), nil
		}
	}
	prefix := "[]"
	if val.Type().Kind() == reflect.Array {
		prefix = fmt.Sprintf("[%d]", val.Len())
	}
	prefix += val.Type().Elem().Kind().String()
	parts := make([]string, val.Len())
	for i := 0; i < val.Len(); i++ {
		var err error
		parts[i], err = p.exporter.export(val.Index(i).Interface())
		if err != nil {
			return "", fmt.Errorf("unexpected err (%s)[%d]: %w", prefix, i, err)
		}
	}
	return prefix + "{" + strings.Join(parts, ", ") + "}", nil
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

type multiArray struct {
	exporter exporter
}

func (m multiArray) export(v any) (string, error) {
	val := reflect.ValueOf(v)
	t := val.Type()
	prefix := ""
	for isBuiltInSliceOrArray(t) {
		if t.Kind() == reflect.Array {
			prefix += fmt.Sprintf("[%d]", t.Len())
		} else {
			prefix += "[]"
		}
		t = t.Elem()
	}

	var ts string
	if t.Kind() == reflect.Interface {
		ts = "interface{}"
	} else {
		ts = t.Kind().String()
	}

	if val.Kind() == reflect.Slice && val.IsNil() {
		return fmt.Sprintf("(%s%s)(nil)", prefix, ts), nil
	}

	parts := make([]string, val.Len())
	for i := 0; i < val.Len(); i++ {
		var err error
		parts[i], err = m.exporter.export(val.Index(i).Interface())
		if err != nil {
			return "", fmt.Errorf("cannot export (%s)[%d]: %w", prefix+ts, i, err)
		}
	}

	return prefix + ts + "{" + strings.Join(parts, ", ") + "}", nil
}

func (m multiArray) supports(v any) bool {
	val := reflect.ValueOf(v)
	if !val.IsValid() {
		return false
	}
	t := val.Type()
	if !isBuiltInSliceOrArray(t) {
		return false
	}
	for isBuiltInSliceOrArray(t) {
		t = t.Elem()
	}

	// workaround: we have to check PkgPath && NumMethod, otherwise
	//
	// z := reflect.Zero(t).Interface()
	// m.exporter.supports(z) // it will return true for interface with methods, e.g. interface{ Do() }
	if t.PkgPath() != "" {
		return false
	}
	if t.Kind() == reflect.Interface && t.NumMethod() > 0 {
		return false
	}

	z := reflect.Zero(t).Interface()
	return m.exporter.supports(z)
}
