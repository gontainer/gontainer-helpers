package reflect

import (
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvert(t *testing.T) {
	// TODO
	t.Run(`Convert parameters`, func(t *testing.T) {
		float64Val := float64(5)
		_ = float64Val // TODO

		a := make([]any, 1) // TODO
		a[0] = a

		b := make([]any, 1)
		b[0] = b

		scenarios := map[string]struct {
			input  any
			output any
			error  string
		}{
			`recursive #1`: {
				input:  a,
				output: b,
			},
			`[]any{[]int{1, 2, 3}} to [][2]int{}`: {
				input:  []any{[]int{1, 2, 3}},
				output: [][2]int{},
				error:  "cannot convert []interface {} to [][2]int: #0: cannot convert []int (length 3) to [2]int",
			},
			`[]any to [0]int`: {
				input:  []any{},
				output: [0]int{},
			},
			`[0]int to []any`: {
				input:  [0]int{},
				output: []any{},
			},
			`[][3]int to [][2]int`: {
				input:  [][3]int{},
				output: [][2]int{},
				error:  `cannot convert [][3]int to [][2]int: cannot convert [3]int to [2]int`,
			},
			`[][3]int to [][3]int`: {
				input:  [][3]int{{5, 5, 5}, {6, 6, 6}},
				output: [][3]int{{5, 5, 5}, {6, 6, 6}},
			},
			`[][3]int to [][3]uint`: {
				input:  [][3]int{{1, 2, 3}},
				output: [][3]uint{{1, 2, 3}},
			},
			`[][3]int to [][3]any`: {
				input:  [][3]int{{2, 2, 2}},
				output: [][3]any{{2, 2, 2}},
			},
			`[][3]any to [][3]int`: {
				input:  [][3]any{{3, 5, 7}},
				output: [][3]int{{3, 5, 7}},
			},
			`[]any{[2]int{}} to [][3]int error`: {
				input:  []any{[2]int{5, 5}},
				output: [][3]int{{5, 5, 0}},
			},
			`[][]any to [][]int`: {
				input:  [][]any{{1, 2, 3}},
				output: [][]int{{1, 2, 3}},
			},
			`[][]any to [][]int (invalid)`: {
				input:  [][]any{{1, false, 3}},
				output: [][]int{{1, 2, 3}},
				error:  "cannot convert [][]interface {} to [][]int: #0: cannot convert []interface {} to []int: #1: cannot convert bool to int",
			},
			`[][]int to [][]any`: {
				input:  [][]int{{1, 2, 3}},
				output: [][]any{{1, 2, 3}},
			},
			`[][]uint to [][]int`: {
				input:  [][]uint{{1, 2, 3}},
				output: [][]int{{1, 2, 3}},
			},
			`[]any to []int`: {
				input:  []any{1, 2, 3},
				output: []int{1, 2, 3},
			},
			`[]any to []int (invalid #1)`: {
				input:  []any{1, 2, nil},
				output: []int{},
				error:  "cannot convert []interface {} to []int: #2: cannot convert <nil> to int",
			},
			`[]any to []int (invalid #2)`: {
				input:  []any{1, 2, 3, struct{}{}},
				output: []int{},
				error:  "cannot convert []interface {} to []int: #3: cannot convert struct {} to int",
			},
			`[]any to []*int`: {
				input:  []any{nil, nil},
				output: []*int{nil, nil},
			},
			`[]int to []any`: {
				input:  []int{1, 2, 3},
				output: []any{1, 2, 3},
			},
			`[]any to []any`: {
				input:  []any{1, 2, 3},
				output: []any{1, 2, 3},
			},
			`[]int8 to []int`: {
				input:  []int8{1, 2, 3},
				output: []int{1, 2, 3},
			},
			`[]int to []int8`: {
				input:  []int{1, 2, 256},
				output: []int8{1, 2, 0},
			},
			`[]struct{}{} to []type`: {
				input:  []struct{}{},
				output: []int{},
				error:  `cannot convert []struct {} to []int: cannot convert struct {} to int`,
			},
			`float64 to int`: {
				input:  float64(math.Pi),
				output: 3,
			},
			`nil to int`: {
				input:  nil,
				output: 0,
				error:  "cannot convert <nil> to int",
			},
			`nil to *int`: {
				input:  nil,
				output: (*int)(nil),
			},
			`*float64 to *int`: {
				input:  &float64Val,
				output: (*int)(nil),
				error:  "cannot convert *float64 to *int",
			},
			`*float64 to *float32`: {
				input:  &float64Val,
				output: (*float32)(nil),
				error:  "cannot convert *float64 to *float32",
			},
			`*float64 to *float64`: {
				input:  &float64Val,
				output: &float64Val,
			},
			`int to float64`: {
				input:  int(5),
				output: float64(5),
			},
			`string to []byte`: {
				input:  `hello`,
				output: []byte(`hello`),
			},
			`[]byte to string`: {
				input:  []byte(`hello`),
				output: `hello`,
			},
			`string to int`: { // cannot convert string to int
				input:  `5`,
				output: int(5),
				error:  "cannot convert string to int",
			},
			`int to string`: { // but reverse conversion is possible, isn't worth to unify this behavior?
				input:  5,
				output: "\x05",
			},
			`nil to *any`: {
				input:  nil,
				output: (*any)(nil),
			},
		}

		for n, tmp := range scenarios {
			s := tmp
			t.Run(n, func(t *testing.T) {
				v, err := convert(s.input, reflect.TypeOf(s.output))
				if s.error != "" {
					assert.EqualError(t, err, s.error, v)
					return
				}

				require.NoError(t, err)
				assert.True(t, reflect.DeepEqual(s.output, v.Interface()))
			})
		}
	})
}
