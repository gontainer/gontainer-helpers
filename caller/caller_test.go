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

package caller_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/gontainer/gontainer-helpers/v2/caller"
	errAssert "github.com/gontainer/gontainer-helpers/v2/grouperror/assert"
	"github.com/stretchr/testify/assert"
)

func TestCall(t *testing.T) {
	t.Run("Given method", func(t *testing.T) {
		p := person{}
		assert.Equal(t, "", p.name)
		_, err := caller.Call(p.setName, []any{"Mary"}, false)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "Mary", p.name)
	})

	t.Run("Given invalid functions", func(t *testing.T) {
		scenarios := []struct {
			fn any
		}{
			{fn: 5},
			{fn: false},
			{fn: (*error)(nil)},
			{fn: struct{}{}},
		}
		const expectedRegexp = "\\Acannot call .*: expected func, .* given\\z"
		for i, tmp := range scenarios {
			s := tmp
			t.Run(fmt.Sprintf("Scenario #%d", i), func(t *testing.T) {
				t.Parallel()
				_, err := caller.Call(s.fn, nil, false)
				assert.Error(t, err)
				assert.Regexp(t, expectedRegexp, err)
			})
		}
	})

	t.Run("Given invalid argument", func(t *testing.T) {
		const msg = "cannot call func([]int): arg0: cannot convert struct {} to []int"
		callee := func([]int) {}
		params := []any{
			struct{}{},
		}

		_, err := caller.Call(callee, params, true)
		assert.EqualError(t, err, msg)
	})

	t.Run("Given invalid arguments", func(t *testing.T) {
		callee := func([]int, *int) {}
		params := []any{
			struct{}{},
			"*int",
		}

		_, err := caller.Call(callee, params, true)

		expected := []string{
			"cannot call func([]int, *int): arg0: cannot convert struct {} to []int",
			"cannot call func([]int, *int): arg1: cannot convert string to *int",
		}
		errAssert.EqualErrorGroup(t, err, expected)
	})

	t.Run("Given too many arguments", func(t *testing.T) {
		const msg = "too many input arguments"
		scenarios := []struct {
			fn   any
			args []any
		}{
			{
				fn:   strings.Join,
				args: []any{"1", "2", "3"},
			},
			{
				fn:   func() {},
				args: []any{1},
			},
		}
		for i, tmp := range scenarios {
			s := tmp
			t.Run(fmt.Sprintf("Scenario #%d", i), func(t *testing.T) {
				t.Parallel()
				_, err := caller.Call(s.fn, s.args, true)
				assert.ErrorContains(t, err, msg)
			})
		}
	})

	t.Run("Given too few arguments", func(t *testing.T) {
		const msg = "not enough input arguments"
		scenarios := []struct {
			fn   any
			args []any
		}{
			{
				fn:   strings.Join,
				args: []any{},
			},
			{
				fn:   func(a int) {},
				args: []any{},
			},
		}
		for i, tmp := range scenarios {
			s := tmp
			t.Run(fmt.Sprintf("Scenario #%d", i), func(t *testing.T) {
				t.Parallel()
				_, err := caller.Call(s.fn, s.args, true)
				assert.ErrorContains(t, err, msg)
			})
		}
	})

	t.Run("Given scenarios", func(t *testing.T) {
		scenarios := []struct {
			fn       any
			args     []any
			expected []any
		}{
			{
				fn: func(a, b int) int {
					return a + b
				},
				args:     []any{uint(1), uint(2)},
				expected: []any{int(3)},
			},
			{
				fn: func(a, b uint) uint {
					return a + b
				},
				args:     []any{int(7), int(3)},
				expected: []any{uint(10)},
			},
			{
				fn: func(vals ...uint) (result uint) {
					for _, v := range vals {
						result += v
					}
					return
				},
				args:     []any{int(1), int(2), int(3)},
				expected: []any{uint(6)},
			},
		}
		for i, tmp := range scenarios {
			s := tmp
			t.Run(fmt.Sprintf("Scenario #%d", i), func(t *testing.T) {
				t.Parallel()
				r, err := caller.Call(s.fn, s.args, true)
				assert.NoError(t, err)
				assert.Equal(t, s.expected, r)
			})
		}
	})

	t.Run("Convert parameters", func(t *testing.T) {
		scenarios := map[string]struct {
			fn     any
			input  any
			output any
			error  string
		}{
			"[]any to []type": {
				fn: func(v []int) []int {
					return v
				},
				input:  []any{1, 2, 3},
				output: []int{1, 2, 3},
			},
			"[]struct{}{} to []type": {
				fn:    func([]int) {},
				input: []struct{}{},
				error: `cannot call func([]int): arg0: cannot convert []struct {} to []int: cannot convert struct {} to int`,
			},
			"nil to any": {
				fn: func(v any) any {
					return v
				},
				input:  nil,
				output: (any)(nil),
			},
		}

		for n, tmp := range scenarios {
			s := tmp
			t.Run(n, func(t *testing.T) {
				t.Parallel()
				r, err := caller.Call(s.fn, []any{s.input}, true)
				if s.error != "" {
					assert.EqualError(t, err, s.error)
					assert.Nil(t, r)
					return
				}

				assert.NoError(t, err)
				assert.Equal(t, r[0], s.output)
			})
		}
	})
}

func TestCallProvider(t *testing.T) {
	t.Run("Given scenarios", func(t *testing.T) {
		scenarios := []struct {
			provider any
			params   []any
			expected any
		}{
			{
				provider: func() any {
					return nil
				},
				expected: nil,
			},
			{
				provider: func(vals ...int) (int, error) {
					result := 0
					for _, v := range vals {
						result += v
					}

					return result, nil
				},
				params:   []any{10, 100, 200},
				expected: 310,
			},
		}

		for i, tmp := range scenarios {
			s := tmp
			t.Run(fmt.Sprintf("Scenario #%d", i), func(t *testing.T) {
				t.Parallel()
				r, err := caller.CallProvider(s.provider, s.params, false)
				assert.NoError(t, err)
				assert.Equal(t, s.expected, r)
			})
		}
	})

	t.Run("Given errors", func(t *testing.T) {
		scenarios := []struct {
			provider any
			params   []any
			err      string
		}{
			{
				provider: func() {},
				err:      "cannot call provider func(): provider must return 1 or 2 values, given function returns 0 values",
			},
			{
				provider: func() (any, any, any) {
					return nil, nil, nil
				},
				err: "cannot call provider func() (interface {}, interface {}, interface {}): provider must return 1 or 2 values, given function returns 3 values",
			},
			{
				provider: func() (any, any) {
					return nil, nil
				},
				err: "cannot call provider func() (interface {}, interface {}): second value returned by provider must implement error interface",
			},
			{
				provider: func() (any, error) {
					return nil, errors.New("test error")
				},
				err: "cannot call provider func() (interface {}, error): test error", // todo maybe provider returned an error?
			},
			{
				provider: func() any {
					return nil
				},
				params: []any{1, 2, 3},
				err:    "cannot call provider func() interface {}: too many input arguments",
			},
		}

		for i, tmp := range scenarios {
			s := tmp
			t.Run(fmt.Sprintf("Scenario #%d", i), func(t *testing.T) {
				t.Parallel()
				r, err := caller.CallProvider(s.provider, s.params, false)
				assert.Nil(t, r)
				assert.EqualError(t, err, s.err)
			})
		}
	})

	t.Run("Given invalid provider", func(t *testing.T) {
		_, err := caller.CallProvider(5, nil, false)
		assert.EqualError(t, err, "cannot call provider int: expected func, int given")
	})

	t.Run("Given provider panics", func(t *testing.T) {
		defer func() {
			assert.Equal(t, "panic!", recover())
		}()

		_, _ = caller.CallProvider(
			func() any {
				panic("panic!")
			},
			nil,
			false,
		)
	})
}

func TestCallWitherByName(t *testing.T) {
	t.Run("Given scenarios", func(t *testing.T) {
		var emptyPerson any = person{}

		scenarios := []struct {
			object any
			wither string
			params []any
			output any
		}{
			{
				object: make(ints, 0),
				wither: "Append",
				params: []any{5},
				output: ints{5},
			},
			{
				object: person{name: "Mary"},
				wither: "WithName",
				params: []any{"Jane"},
				output: person{name: "Jane"},
			},
			{
				object: &person{name: "Mary"},
				wither: "WithName",
				params: []any{"Jane"},
				output: person{name: "Jane"},
			},
			{
				object: emptyPerson,
				wither: "WithName",
				params: []any{"Kaladin"},
				output: person{name: "Kaladin"},
			},
			{
				object: &emptyPerson,
				wither: "WithName",
				params: []any{"Shallan"},
				output: person{name: "Shallan"},
			},
		}

		for i, tmp := range scenarios {
			s := tmp
			t.Run(fmt.Sprintf("Scenario #%d", i), func(t *testing.T) {
				t.Parallel()
				result, err := caller.CallWitherByName(s.object, s.wither, s.params, false)
				assert.NoError(t, err)
				assert.Equal(t, s.output, result)
			})
		}
	})

	t.Run("Given errors", func(t *testing.T) {
		scenarios := []struct {
			object any
			wither string
			params []any
			error  string
		}{
			{
				object: person{},
				wither: "withName",
				params: nil,
				error:  `cannot call wither (caller_test.person)."withName": invalid func (caller_test.person)."withName"`,
			},
			{
				object: person{},
				wither: "Clone",
				params: nil,
				error:  `cannot call wither (caller_test.person)."Clone": wither must return 1 value, given function returns 2 values`,
			},
			{
				object: person{},
				wither: "WithName",
				params: nil,
				error:  `cannot call wither (caller_test.person)."WithName": not enough input arguments`,
			},
		}

		for i, tmp := range scenarios {
			s := tmp
			t.Run(fmt.Sprintf("Scenario #%d", i), func(t *testing.T) {
				t.Parallel()
				o, err := caller.CallWitherByName(s.object, s.wither, s.params, false)
				assert.Nil(t, o)
				assert.EqualError(t, err, s.error)
			})
		}
	})
}

type ints []int

func (i ints) Append(v int) ints {
	return append(i, v)
}

type person struct {
	name string
}

func (p person) Clone() (person, error) {
	return p, nil
}

func (p person) WithName(n string) person {
	return person{name: n}
}

func (p person) withName(n string) person { //nolint:unused
	return person{name: n}
}

func (p *person) setName(n string) {
	p.name = n
}
