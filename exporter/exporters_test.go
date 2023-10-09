package exporter

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

type myString string
type aliasString = string
type myInt int
type aliasInt = int
type myBool bool
type aliasBool = bool

type mockExporter struct {
	result string
	error  error
}

func (m mockExporter) export(interface{}) (string, error) {
	return m.result, m.error
}

func TestChainExporter_Export(t *testing.T) {
	exporter := newDefaultExporter()

	t.Run("Given scenarios", func(t *testing.T) {
		scenarios := map[string]struct {
			input  interface{}
			output string
			error  string
		}{
			"nil": {
				input:  nil,
				output: "nil",
			},
			"false": {
				input:  false,
				output: "false",
			},
			"true": {
				input:  true,
				output: "true",
			},
			"123": {
				input:  int(123),
				output: "int(123)",
			},
			"`hello world`": {
				input:  "hello world",
				output: `"hello world"`,
			},
			"[]byte": {
				input:  []byte("hello world 你好，世界"),
				output: `[]byte("hello world \u4f60\u597d\uff0c\u4e16\u754c")`,
			},
			"struct {}": {
				input: struct{}{},
				error: "type `struct {}` is not supported",
			},
			"*testing.T": {
				input: t,
				error: "type `*testing.T` is not supported",
			},
			`myString("foo")`: {
				input: myString("foo"),
				error: "type `exporter.myString` is not supported",
			},
			`aliasString("foo")`: {
				input:  aliasString("foo"),
				output: `"foo"`,
			},
			`myInt(5)`: {
				input: myInt(5),
				error: "type `exporter.myInt` is not supported",
			},
			`aliasInt(5)`: {
				input:  aliasInt(5),
				output: "int(5)",
			},
			`myBool(true)`: {
				input: myBool(true),
				error: "type `exporter.myBool` is not supported",
			},
			`aliasBool(true)`: {
				input:  aliasBool(true),
				output: "true",
			},
		}

		for k, tmp := range scenarios {
			s := tmp
			t.Run(k, func(t *testing.T) {
				t.Parallel()
				output, err := exporter.export(s.input)
				if s.error != "" {
					assert.EqualError(t, err, s.error)
					assert.Equal(t, "", output)
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, s.output, output)
			})
		}
	})
}

func TestExport(t *testing.T) {
	scenarios := []struct {
		input  interface{}
		output string
		error  string
		panic  string
	}{
		{
			input:  123,
			output: "int(123)",
		},
		{
			input:  []interface{}{1, "2", 3.14},
			output: `[]interface{}{int(1), "2", float64(3.14)}`,
		},
		{
			input:  [3]interface{}{1, "2", 3.14},
			output: `[3]interface{}{int(1), "2", float64(3.14)}`,
		},
		{
			input:  []interface{}{},
			output: "make([]interface{}, 0)",
		},
		{
			input:  [0]interface{}{},
			output: "[0]interface{}{}",
		},
		{
			input: []interface{}{struct{}{}},
			error: "cannot export slice[0]: type `struct {}` is not supported",
			panic: "cannot export `[]interface {}` to string: cannot export slice[0]: type `struct {}` is not supported",
		},
		{
			input: [1]interface{}{struct{}{}},
			error: "cannot export array[0]: type `struct {}` is not supported",
			panic: "cannot export `[1]interface {}` to string: cannot export array[0]: type `struct {}` is not supported",
		},
		{
			input:  []int{1, 2, 3},
			output: "[]int{int(1), int(2), int(3)}",
		},
		{
			input:  [3]int{1, 2, 3},
			output: "[3]int{int(1), int(2), int(3)}",
		},
		{
			input:  [0]int{},
			output: "[0]int{}",
		},
		{
			input:  []float32{},
			output: "make([]float32, 0)",
		},
		{
			input:  [0]float32{},
			output: "[0]float32{}",
		},
		{
			input: struct{}{},
			error: "type `struct {}` is not supported",
			panic: "cannot export `struct {}` to string: type `struct {}` is not supported",
		},
		{
			input: []interface{ Do() }{nil, nil, nil},
			error: "type `[]interface { Do() }` is not supported",
			panic: "cannot export `[]interface { Do() }` to string: type `[]interface { Do() }` is not supported",
		},
		{
			input: [3]interface{ Do() }{},
			error: "type `[3]interface { Do() }` is not supported",
			panic: "cannot export `[3]interface { Do() }` to string: type `[3]interface { Do() }` is not supported",
		},
		{
			input:  []interface{}{nil, nil, nil},
			output: "[]interface{}{nil, nil, nil}",
		},
		{
			input:  [3]interface{}{},
			output: "[3]interface{}{nil, nil, nil}",
		},
	}

	for i, tmp := range scenarios {
		s := tmp
		t.Run(fmt.Sprintf("Scenario #%d", i), func(t *testing.T) {
			t.Parallel()
			func() {
				defer func() {
					r := recover()
					if s.panic == "" {
						assert.Nil(t, r)
						return
					}
					assert.Equal(t, s.panic, r)
				}()
				assert.Equal(t, s.output, MustExport(s.input))
			}()

			o, err := Export(s.input)
			if s.error == "" {
				assert.NoError(t, err)
				assert.Equal(t, s.output, o)
				return
			}

			assert.EqualError(t, err, s.error)
		})
	}

	t.Run("Given invalid scenario", func(t *testing.T) {
		originalExporter := defaultExporter
		defer func() {
			defaultExporter = originalExporter
		}()

		expectedErr := fmt.Errorf("my test error")
		defaultExporter = mockExporter{
			error: expectedErr,
		}
		_, err := Export(123)
		assert.EqualError(t, err, expectedErr.Error())
	})
}

func TestToString(t *testing.T) {
	scenarios := []struct {
		input  interface{}
		output string
		error  string
	}{
		{
			input:  true,
			output: "true",
		},
		{
			input:  nil,
			output: "nil",
		},
		{
			input: struct{}{},
			error: "type `struct {}` is not supported",
		},
		{
			input:  "Mary Jane",
			output: "Mary Jane",
		},
		{
			input:  int(5),
			output: "5",
		},
		{
			input:  float64(3.14),
			output: "3.14",
		},
	}

	for i, tmp := range scenarios {
		s := tmp
		t.Run(fmt.Sprintf("Scenario #%d", i), func(t *testing.T) {
			t.Parallel()
			t.Run("ToString", func(t *testing.T) {
				result, err := ToString(s.input)

				if s.error != "" {
					assert.Empty(t, result)
					assert.EqualError(t, err, s.error)
					return
				}

				assert.NoError(t, err)
				assert.Equal(t, s.output, result)
			})

			t.Run("MustToString", func(t *testing.T) {
				defer func() {
					err := recover()
					if s.error == "" {
						assert.Nil(t, err)
						return
					}

					assert.NotNil(t, err)

					assert.Equal(
						t,
						fmt.Sprintf(
							"cannot cast `%T` to string: %s",
							s.input,
							s.error,
						),
						fmt.Sprintf("%s", err),
					)
				}()

				assert.Equal(t, s.output, MustToString(s.input))
			})
		})
	}
}

func TestNumericExporter_Supports(t *testing.T) {
	scenarios := []struct {
		input    interface{}
		expected bool
	}{
		{
			input:    nil,
			expected: false,
		},
		{
			input:    math.Pi,
			expected: true,
		},
		{
			input:    "3.14",
			expected: false,
		},
	}

	for i, tmp := range scenarios {
		s := tmp
		t.Run(fmt.Sprintf("Scenario #%d", i), func(t *testing.T) {
			t.Parallel()
			assert.Equal(
				t,
				s.expected,
				numberExporter{}.supports(s.input),
			)
		})
	}
}

func TestPrimitiveTypeSliceExporter_Supports(t *testing.T) {
	scenarios := []struct {
		input    interface{}
		expected bool
	}{
		{
			input:    nil,
			expected: false,
		},
		{
			input:    math.Pi,
			expected: false,
		},
		{
			input:    "3.14",
			expected: false,
		},
		{
			input:    []uint{0, 1},
			expected: true,
		},
		{
			input:    []struct{}{},
			expected: false,
		},
		{
			input:    []myBool{},
			expected: false,
		},
	}

	for i, tmp := range scenarios {
		s := tmp
		t.Run(fmt.Sprintf("Scenario #%d", i), func(t *testing.T) {
			t.Parallel()
			assert.Equal(
				t,
				s.expected,
				primitiveTypeSliceExporter{}.supports(s.input),
			)
		})
	}
}

func TestInterfaceSliceExporter_Supports(t *testing.T) {
	scenarios := []struct {
		input    interface{}
		expected bool
	}{
		{
			input:    nil,
			expected: false,
		},
		{
			input:    math.Pi,
			expected: false,
		},
		{
			input:    "3.14",
			expected: false,
		},
		{
			input:    []uint{0, 1},
			expected: false,
		},
		{
			input:    []struct{}{},
			expected: false,
		},
		{
			input:    []interface{}{},
			expected: true,
		},
		{
			input:    []myBool{},
			expected: false,
		},
	}

	for i, tmp := range scenarios {
		s := tmp
		t.Run(fmt.Sprintf("Scenario #%d", i), func(t *testing.T) {
			t.Parallel()
			assert.Equal(
				t,
				s.expected,
				interfaceSliceExporter{}.supports(s.input),
				fmt.Sprintf("value: %#v", s.input),
			)
		})
	}
}

func TestPrimitiveTypeSliceExporter_Export(t *testing.T) {
	t.Run("Given error in subexporter", func(t *testing.T) {
		exp := primitiveTypeSliceExporter{
			exporter: chainExporter{},
		}
		v, err := exp.export([]uint{1})
		assert.Equal(t, "", v)
		assert.EqualError(
			t,
			err,
			"unexpected err slice[0]: type `uint` is not supported",
		)
	})
}
