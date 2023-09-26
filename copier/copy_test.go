package copier_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/gontainer/gontainer-helpers/copier"
	"github.com/stretchr/testify/assert"
)

func TestCopy(t *testing.T) {
	t.Run("var from interface{}", func(t *testing.T) {
		var from interface{} = car{age: 5}
		var to car

		assert.NoError(t, copier.Copy(from, &to))
		assert.Equal(t, 5, to.age)
	})
	t.Run("var to interface{}", func(t *testing.T) {
		six := 6
		scenarios := []interface{}{
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
				var to interface{}
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
