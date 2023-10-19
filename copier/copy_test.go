package copier_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gontainer/gontainer-helpers/copier"
	"github.com/stretchr/testify/assert"
)

func TestCopy(t *testing.T) {
	t.Run("Convert value", func(t *testing.T) {
		t.Run("Copy", func(t *testing.T) {
			var (
				from int = 5
				to   uint
			)
			err := copier.Copy(from, &to)
			assert.Empty(t, to)
			assert.EqualError(t, err, "value of type int is not assignable to type uint")
		})
		t.Run("ConvertAndCopy", func(t *testing.T) {
			var (
				from int = 5
				to   uint
			)
			err := copier.ConvertAndCopy(from, &to)
			assert.Equal(t, uint(5), to)
			assert.NoError(t, err)
		})
	})
	t.Run("Convert pointer", func(t *testing.T) {
		t.Run("Copy", func(t *testing.T) {
			var (
				from *int
				to   *uint
			)
			err := copier.Copy(from, &to)
			assert.Empty(t, to)
			assert.EqualError(t, err, "value of type *int is not assignable to type *uint")
		})
		t.Run("ConvertAndCopy", func(t *testing.T) {
			var (
				from *int
				to   *uint
			)
			err := copier.ConvertAndCopy(from, &to)
			assert.Empty(t, to)
			assert.EqualError(t, err, "cannot cast `*int` to `*uint`")
		})
	})
	t.Run("ConvertAndCopy non-empty interface", func(t *testing.T) {
		t.Run("Copy", func(t *testing.T) {
			var (
				from interface{ Foo() }
				to   interface{ Bar() } // even tho interfaces differ, it does not return an error
			)
			err := copier.Copy(from, &to)
			assert.Empty(t, to)
			assert.NoError(t, err)
		})
		t.Run("ConvertAndCopy", func(t *testing.T) {
			var (
				from interface{ Foo() }
				to   interface{ Bar() } // even tho interfaces differ, it does not return an error
			)
			err := copier.ConvertAndCopy(from, &to)
			assert.Empty(t, to)
			assert.NoError(t, err)
		})
	})
	t.Run("var from any", func(t *testing.T) {
		var from any = car{age: 5}
		var to car

		assert.NoError(t, copier.Copy(from, &to))
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
				assert.NoError(t, copier.Copy(d, &to))
				assert.Equal(t, d, to)
				if reflect.ValueOf(d).Kind() == reflect.Ptr {
					assert.Same(t, d, to)
				}
			})
		}
	})
	t.Run("Given errors", func(t *testing.T) {
		t.Run("non-pointer value", func(t *testing.T) {
			const msg = "expected pointer, `int` given"
			assert.EqualError(
				t,
				copier.Copy(5, 5),
				msg,
			)
		})
	})
}

type car struct {
	age int
}
