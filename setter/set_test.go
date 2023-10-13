package setter

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_kindChain_isInterfaceOverPointerChain(t *testing.T) {
	scenarios := []struct {
		chain    kindChain
		expected bool
	}{
		{
			chain:    nil,
			expected: false,
		},
		{
			chain:    kindChain{reflect.Struct, reflect.Struct, reflect.Struct, reflect.Struct},
			expected: false,
		},
		{
			chain:    kindChain{reflect.Ptr, reflect.Struct, reflect.Struct, reflect.Interface},
			expected: false,
		},
		{
			chain:    kindChain{reflect.Ptr, reflect.Interface, reflect.Struct, reflect.Interface},
			expected: false,
		},
		{
			chain:    kindChain{reflect.Ptr, reflect.Interface, reflect.Struct, reflect.Struct},
			expected: false,
		},
		{
			chain:    kindChain{reflect.Ptr, reflect.Interface, reflect.Ptr, reflect.Struct},
			expected: true,
		},
		{
			chain:    kindChain{reflect.Ptr, reflect.Interface, reflect.Ptr, reflect.Ptr, reflect.Struct},
			expected: true,
		},
	}

	for i := 3; i < 10; i++ {
		ptrs := make(kindChain, 0)
		for j := 0; j < i; j++ {
			ptrs = append(ptrs, reflect.Ptr)
		}

		chain := kindChain{reflect.Ptr, reflect.Interface}
		chain = append(chain, ptrs...)
		chain = append(chain, reflect.Struct)

		scenarios = append(scenarios, struct {
			chain    kindChain
			expected bool
		}{
			chain:    chain,
			expected: true,
		})
	}

	for id, tmp := range scenarios {
		s := tmp
		t.Run(fmt.Sprintf("scenario %d", id), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, s.expected, s.chain.isInterfaceOverPointerChain())
		})
	}
}

func TestSet(t *testing.T) {
	t.Run("blank identifier", func(t *testing.T) {
		assert.EqualError(
			t,
			Set(struct{ _ int }{}, "_", 123),
			`"_" is not supported`,
		)
	})
	t.Run("anonymous struct", func(t *testing.T) {
		p := struct {
			color string
		}{}
		assert.NoError(t, Set(&p, "color", "red"))
		assert.Equal(t, "red", p.color)
	})
	t.Run("anonymous *struct", func(t *testing.T) {
		p := &struct {
			color string
		}{}
		assert.NoError(t, Set(&p, "color", "brown"))
		assert.Equal(t, "brown", p.color)
	})
	t.Run("***struct", func(t *testing.T) {
		p := &struct {
			color string
		}{}
		p2 := &p
		p3 := &p2
		assert.NoError(t, Set(&p3, "color", "brown"))
		assert.Equal(t, "brown", p.color)
	})
	t.Run("var a interface{}", func(t *testing.T) {
		t.Run("*struct{}", func(t *testing.T) {
			const color = "red"
			p := struct {
				color string
			}{}
			var obj interface{} = &p
			assert.Equal(t, "", p.color)
			assert.NoError(t, Set(obj, "color", color))
			assert.Equal(t, color, p.color)
		})
		t.Run("**struct{}", func(t *testing.T) {
			const color = "blue"
			p := struct {
				color string
			}{}
			p2 := &p
			var obj interface{} = &p2
			assert.Equal(t, "", p.color)
			assert.NoError(t, Set(obj, "color", color))
			assert.Equal(t, color, p.color)
		})
		t.Run("***struct{}", func(t *testing.T) {
			const color = "yellow"
			p := struct {
				color string
			}{}
			p2 := &p
			p3 := &p2
			var obj interface{} = &p3
			assert.Equal(t, "", p.color)
			assert.NoError(t, Set(obj, "color", color))
			assert.Equal(t, color, p.color)
		})
		t.Run("****struct{}", func(t *testing.T) {
			const color = "green"
			p := struct {
				color string
			}{}
			p2 := &p
			p3 := &p2
			p4 := &p3
			var obj interface{} = &p4
			assert.Equal(t, "", p.color)
			assert.NoError(t, Set(obj, "color", color))
			assert.Equal(t, color, p.color)
		})
	})
	t.Run("struct", func(t *testing.T) {
		p := person{}
		assert.NoError(t, Set(&p, "Name", "Jane"))
		assert.NoError(t, Set(&p, "age", 30))
		assert.Equal(t, person{Name: "Jane", age: 30}, p)
	})
	t.Run("*struct", func(t *testing.T) {
		p := &person{}
		assert.NoError(t, Set(&p, "Name", "Mary"))
		assert.NoError(t, Set(&p, "age", uint(33)))
		assert.Equal(t, &person{Name: "Mary", age: 33}, p)
	})
	t.Run("var a interface{} = &struct{}", func(t *testing.T) {
		var p interface{} = &person{}
		assert.NoError(t, Set(&p, "Name", "Mary Jane"))
		assert.NoError(t, Set(&p, "age", 45))
		assert.Equal(t, &person{Name: "Mary Jane", age: 45}, p)
	})
	t.Run("var a interface{} = struct{}", func(t *testing.T) {
		var p interface{} = person{}
		assert.NoError(t, Set(&p, "Name", "Jane"))
		assert.Equal(t, person{Name: "Jane"}, p)
	})
	t.Run("unexported type of field", func(t *testing.T) {
		p := person{}
		assert.NoError(t, Set(&p, "wallet", wallet{amount: 400}))
		assert.Equal(t, wallet{amount: 400}, p.wallet)
	})
	t.Run("convert []interface{} to []type", func(t *testing.T) {
		s := storage{}
		assert.NoError(
			t,
			Set(&s, "wallets", []interface{}{wallet{100}, wallet{200}}),
		)
		assert.Equal(
			t,
			[]wallet{{100}, {200}},
			s.wallets,
		)
	})
	t.Run("convert int16 to uint", func(t *testing.T) {
		var p struct {
			Age uint
		}
		err := Set(&p, "Age", int16(20))
		assert.NoError(t, err)
		assert.Equal(t, uint(20), p.Age)
	})
	t.Run("Given errors", func(t *testing.T) {
		t.Run("Field does not exist", func(t *testing.T) {
			p := person{}
			err := Set(&p, "FirstName", "Mary")
			assert.EqualError(t, err, "set `*setter.person`.\"FirstName\": field `FirstName` does not exist")
		})
		t.Run("Invalid pointer dest", func(t *testing.T) {
			p := 5
			err := Set(&p, "FirstName", "Mary")
			assert.EqualError(t, err, "expected pointer to struct, ptr.int given")
		})
		t.Run("Invalid type of value", func(t *testing.T) {
			p := person{}
			err := Set(&p, "Name", struct{}{})
			assert.EqualError(t, err, "set `*setter.person`.\"Name\": cannot cast `struct {}` to `string`")
		})
		t.Run("Invalid type of value (var p interface{} = person{})", func(t *testing.T) {
			var p interface{} = person{}
			err := Set(&p, "Name", struct{}{})
			assert.EqualError(t, err, "set `*interface {}`.\"Name\": cannot cast `struct {}` to `string`")
		})
	})
}

type person struct {
	Name   string
	age    uint8
	wallet wallet
}

type wallet struct {
	amount uint
}

type storage struct {
	wallets []wallet
}
