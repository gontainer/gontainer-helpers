package copier_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gontainer/gontainer-helpers/v2/copier"
	"github.com/stretchr/testify/assert"
)

func TestCopy(t *testing.T) {
	t.Run("Convert value", func(t *testing.T) {
		t.Run("Do not convert", func(t *testing.T) {
			var (
				from int = 5
				to   uint
			)
			err := copier.Copy(from, &to, false)
			assert.Empty(t, to)
			assert.EqualError(t, err, "value of type int is not assignable to type uint")
		})
		t.Run("Convert", func(t *testing.T) {
			var (
				from int = 5
				to   uint
			)
			err := copier.Copy(from, &to, true)
			assert.Equal(t, uint(5), to)
			assert.NoError(t, err)
		})
	})
	t.Run("Convert pointer", func(t *testing.T) {
		t.Run("Do not convert", func(t *testing.T) {
			var (
				from *int
				to   *uint
			)
			err := copier.Copy(from, &to, false)
			assert.Empty(t, to)
			assert.EqualError(t, err, "value of type *int is not assignable to type *uint")
		})
		t.Run("Convert", func(t *testing.T) {
			var (
				from *int
				to   *uint
			)
			err := copier.Copy(from, &to, true)
			assert.Empty(t, to)
			assert.EqualError(t, err, "cannot convert *int to *uint")
		})
	})
	t.Run("Non-empty interface", func(t *testing.T) {
		t.Run("Do not convert", func(t *testing.T) {
			var (
				from interface{ Foo() }
				to   interface{ Bar() } // even tho interfaces differ, it does not return an error
			)
			err := copier.Copy(from, &to, false)
			assert.Empty(t, to)
			assert.NoError(t, err)
		})
		t.Run("Convert", func(t *testing.T) {
			var (
				from interface{ Foo() }
				to   interface{ Bar() } // even tho interfaces differ, it does not return an error
			)
			err := copier.Copy(from, &to, true)
			assert.Empty(t, to)
			assert.NoError(t, err)
		})
	})
	t.Run("var from any", func(t *testing.T) {
		var from any = car{age: 5}
		var to car

		assert.NoError(t, copier.Copy(from, &to, false))
		assert.Equal(t, 5, to.age)
	})
	t.Run("var to any", func(t *testing.T) {
		six := 6
		scenarios := []any{
			5,
			3.14,
			struct{}{},
			nil,
			&six,
			car{age: 10},
			&car{age: 10},
			(*car)(nil),
		}

		for id, tmp := range scenarios {
			d := tmp
			t.Run(fmt.Sprintf("%d: `%T`", id, d), func(t *testing.T) {
				t.Parallel()
				var to any
				assert.NoError(t, copier.Copy(d, &to, false))
				assert.Equal(t, d, to)
				if reflect.ValueOf(d).Kind() == reflect.Ptr {
					assert.Same(t, d, to)
				}
			})
		}
	})
	t.Run("Given errors", func(t *testing.T) {
		t.Run("non-pointer value", func(t *testing.T) {
			const msg = "expected ptr, int given"
			assert.EqualError(
				t,
				copier.Copy(5, 5, false),
				msg,
			)
		})
	})
	t.Run("Copy to nil", func(t *testing.T) {
		assert.EqualError(
			t,
			copier.Copy(5, nil, false),
			"expected ptr, <nil> given", // TODO unify with other errors
		)
	})
	t.Run("Convert", func(t *testing.T) {
		t.Run("[]int to []any", func(t *testing.T) {
			var (
				from = []int{1, 2, 3}
				to   []any
			)

			err := copier.Copy(from, &to, true)
			assert.NoError(t, err)
			assert.Equal(t, []any{1, 2, 3}, to)
		})
		t.Run("[]any to []int", func(t *testing.T) {
			var (
				from = []any{1, 2, 3}
				to   []int
			)

			err := copier.Copy(from, &to, true)
			assert.NoError(t, err)
			assert.Equal(t, []int{1, 2, 3}, to)
		})
		t.Run("[]int to [N]int", func(t *testing.T) {
			var (
				from = []int{1, 2, 3}
				to   [3]int
			)

			err := copier.Copy(from, &to, true)
			assert.NoError(t, err)
			assert.Equal(t, [3]int{1, 2, 3}, to)
		})
		t.Run("[]int to [N]int #2", func(t *testing.T) {
			var (
				from = []int{1, 2, 3}
				to   [2]int
			)

			err := copier.Copy(from, &to, true)
			assert.EqualError(t, err, "cannot convert []int (length 3) to [2]int")
			assert.Empty(t, to)
		})
		t.Run("[N]int to [N-1]int", func(t *testing.T) {
			var (
				from = [3]int{1, 2, 3}
				to   [2]int
			)

			err := copier.Copy(from, &to, true)
			assert.EqualError(t, err, "cannot convert [3]int to [2]int")
			assert.Empty(t, to)
		})
		t.Run("[N]int to [N+1]int", func(t *testing.T) {
			var (
				from = [3]int{1, 2, 3}
				to   [4]int
			)

			err := copier.Copy(from, &to, true)
			assert.NoError(t, err)
			assert.Equal(t, [4]int{1, 2, 3, 0}, to)
		})
		t.Run("[N]any to [N+1]any", func(t *testing.T) {
			var (
				from = [3]any{6, 7, 8}
				to   [4]any
			)

			err := copier.Copy(from, &to, true)
			assert.NoError(t, err)
			assert.Equal(t, [4]any{6, 7, 8, nil}, to)
		})
		t.Run("[N]int to []int", func(t *testing.T) {
			var (
				from = [3]int{1, 2, 3}
				to   []int
			)

			err := copier.Copy(from, &to, true)
			assert.NoError(t, err)
			assert.Equal(t, []int{1, 2, 3}, to)
		})
		t.Run("[N]int to [N]uint", func(t *testing.T) {
			var (
				from = [3]int{1, 2, 3}
				to   [3]uint
			)

			err := copier.Copy(from, &to, true)
			assert.NoError(t, err)
			assert.Equal(t, [3]uint{1, 2, 3}, to)
		})
	})
}

type car struct {
	age int
}
