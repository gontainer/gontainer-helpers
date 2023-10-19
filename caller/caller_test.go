package caller_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/gontainer/gontainer-helpers/caller"
	errAssert "github.com/gontainer/gontainer-helpers/grouperror/assert"
	"github.com/stretchr/testify/assert"
)

func TestCall(t *testing.T) {
	t.Run("Given method", func(t *testing.T) {
		p := person{}
		assert.Equal(t, "", p.name)
		_, err := caller.Call(p.setName, []interface{}{"Mary"}, false)
		if !assert.NoError(t, err) {
			return
		}
		assert.Equal(t, "Mary", p.name)
	})

	t.Run("Given invalid functions", func(t *testing.T) {
		scenarios := []struct {
			fn interface{}
		}{
			{fn: 5},
			{fn: false},
			{fn: (*error)(nil)},
			{fn: struct{}{}},
		}
		const expectedRegexp = "\\Aexpected `func`, .* given\\z"
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
		const msg = "arg0: cannot cast `struct {}` to `[]int`"
		callee := func([]int) {}
		params := []interface{}{
			struct{}{},
		}

		_, err := caller.Call(callee, params, true)
		assert.EqualError(t, err, msg)
	})

	t.Run("Given invalid arguments", func(t *testing.T) {
		callee := func([]int, *int) {}
		params := []interface{}{
			struct{}{},
			"*int",
		}

		_, err := caller.Call(callee, params, true)

		expected := []string{
			"arg0: cannot cast `struct {}` to `[]int`",
			"arg1: cannot cast `string` to `*int`",
		}
		errAssert.EqualErrorGroup(t, err, expected)
	})

	t.Run("Given too many arguments", func(t *testing.T) {
		const msg = "too many input arguments"
		scenarios := []struct {
			fn   interface{}
			args []interface{}
		}{
			{
				fn:   strings.Join,
				args: []interface{}{"1", "2", "3"},
			},
			{
				fn:   func() {},
				args: []interface{}{1},
			},
		}
		for i, tmp := range scenarios {
			s := tmp
			t.Run(fmt.Sprintf("Scenario #%d", i), func(t *testing.T) {
				t.Parallel()
				_, err := caller.Call(s.fn, s.args, true)
				assert.EqualError(t, err, msg)
			})
		}
	})

	t.Run("Given too few arguments", func(t *testing.T) {
		const msg = "too few input arguments"
		scenarios := []struct {
			fn   interface{}
			args []interface{}
		}{
			{
				fn:   strings.Join,
				args: []interface{}{},
			},
			{
				fn:   func(a int) {},
				args: []interface{}{},
			},
		}
		for i, tmp := range scenarios {
			s := tmp
			t.Run(fmt.Sprintf("Scenario #%d", i), func(t *testing.T) {
				t.Parallel()
				_, err := caller.Call(s.fn, s.args, true)
				assert.EqualError(t, err, msg)
			})
		}
	})

	t.Run("Given scenarios", func(t *testing.T) {
		scenarios := []struct {
			fn       interface{}
			args     []interface{}
			expected []interface{}
		}{
			{
				fn: func(a, b int) int {
					return a + b
				},
				args:     []interface{}{uint(1), uint(2)},
				expected: []interface{}{int(3)},
			},
			{
				fn: func(a, b uint) uint {
					return a + b
				},
				args:     []interface{}{int(7), int(3)},
				expected: []interface{}{uint(10)},
			},
			{
				fn: func(vals ...uint) (result uint) {
					for _, v := range vals {
						result += v
					}
					return
				},
				args:     []interface{}{int(1), int(2), int(3)},
				expected: []interface{}{uint(6)},
			},
		}
		for i, tmp := range scenarios {
			s := tmp
			t.Run(fmt.Sprintf("Scenario #%d", i), func(t *testing.T) {
				t.Parallel()
				// TODO it panics for `false`, return an error
				r, err := caller.Call(s.fn, s.args, true)
				assert.NoError(t, err)
				assert.Equal(t, s.expected, r)
			})
		}
	})

	t.Run("Convert parameters", func(t *testing.T) {
		scenarios := map[string]struct {
			fn     interface{}
			input  interface{}
			output interface{}
			error  string
		}{
			"[]interface{} to []type": {
				fn: func(v []int) []int {
					return v
				},
				input:  []interface{}{1, 2, 3},
				output: []int{1, 2, 3},
			},
			"[]struct{}{} to []type": {
				fn:    func([]int) {},
				input: []struct{}{},
				error: "arg0: cannot cast `[]struct {}` to `[]int`",
			},
			"nil to interface{}": {
				fn: func(v interface{}) interface{} {
					return v
				},
				input:  nil,
				output: (interface{})(nil),
			},
		}

		for n, tmp := range scenarios {
			s := tmp
			t.Run(n, func(t *testing.T) {
				t.Parallel()
				r, err := caller.Call(s.fn, []interface{}{s.input}, true)
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
			provider interface{}
			params   []interface{}
			expected interface{}
		}{
			{
				provider: func() interface{} {
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
				params:   []interface{}{10, 100, 200},
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
			provider interface{}
			params   []interface{}
			err      string
		}{
			{
				provider: func() {},
				err:      "provider must return 1 or 2 values, given function returns 0 values",
			},
			{
				provider: func() (interface{}, interface{}, interface{}) {
					return nil, nil, nil
				},
				err: "provider must return 1 or 2 values, given function returns 3 values",
			},
			{
				provider: func() (interface{}, interface{}) {
					return nil, nil
				},
				err: "second value returned by provider must implement error interface",
			},
			{
				provider: func() (interface{}, error) {
					return nil, errors.New("test error")
				},
				err: "test error",
			},
			{
				provider: func() interface{} {
					return nil
				},
				params: []interface{}{1, 2, 3},
				err:    "too many input arguments",
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
		assert.EqualError(t, err, "expected `func`, `int` given")
	})

	t.Run("Given provider panics", func(t *testing.T) {
		defer func() {
			assert.Equal(t, "panic!", recover())
		}()

		_, _ = caller.CallProvider(
			func() interface{} {
				panic("panic!")
			},
			nil,
			false,
		)
	})
}

func TestCallWitherByName(t *testing.T) {
	t.Run("Given scenarios", func(t *testing.T) {
		var emptyPerson interface{} = person{}

		scenarios := []struct {
			object interface{}
			wither string
			params []interface{}
			output interface{}
		}{
			{
				object: make(ints, 0),
				wither: "Append",
				params: []interface{}{5},
				output: ints{5},
			},
			{
				object: person{name: "Mary"},
				wither: "WithName",
				params: []interface{}{"Jane"},
				output: person{name: "Jane"},
			},
			{
				object: &person{name: "Mary"},
				wither: "WithName",
				params: []interface{}{"Jane"},
				output: person{name: "Jane"},
			},
			{
				object: emptyPerson,
				wither: "WithName",
				params: []interface{}{"Kaladin"},
				output: person{name: "Kaladin"},
			},
			{
				object: &emptyPerson,
				wither: "WithName",
				params: []interface{}{"Shallan"},
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
			object interface{}
			wither string
			params []interface{}
			error  string
		}{
			{
				object: person{},
				wither: "withName",
				params: nil,
				error:  "invalid func `caller_test.person`.\"withName\"",
			},
			{
				object: person{},
				wither: "Clone",
				params: nil,
				error:  "wither must return 1 value, given function returns 2 values",
			},
			{
				object: person{},
				wither: "WithName",
				params: nil,
				error:  "`caller_test.person`.\"WithName\": too few input arguments",
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
