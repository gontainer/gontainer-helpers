package setter_test

import (
	"testing"

	"github.com/gontainer/gontainer-helpers/v2/setter"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	t.Run("blank identifier", func(t *testing.T) {
		assert.EqualError(
			t,
			setter.Set(struct{ _ int }{}, "_", 123, false),
			`set (struct { _ int })."_": "_" is not supported`,
		)
	})
	t.Run("anonymous struct", func(t *testing.T) {
		p := struct {
			color string
		}{}
		assert.NoError(t, setter.Set(&p, "color", "red", false))
		assert.Equal(t, "red", p.color)
	})
	t.Run("anonymous *struct", func(t *testing.T) {
		p := &struct {
			color string
		}{}
		assert.NoError(t, setter.Set(&p, "color", "brown", false))
		assert.Equal(t, "brown", p.color)
	})
	t.Run("***struct", func(t *testing.T) {
		p := &struct {
			color string
		}{}
		p2 := &p
		p3 := &p2
		assert.NoError(t, setter.Set(&p3, "color", "brown", false))
		assert.Equal(t, "brown", p.color)
	})
	t.Run("var a any", func(t *testing.T) {
		t.Run("*struct{}", func(t *testing.T) {
			const color = "red"
			p := struct {
				color string
			}{}
			var obj any = &p
			assert.Equal(t, "", p.color)
			assert.NoError(t, setter.Set(obj, "color", color, false))
			assert.Equal(t, color, p.color)
		})
		t.Run("**struct{}", func(t *testing.T) {
			const color = "blue"
			p := struct {
				color string
			}{}
			p2 := &p
			var obj any = &p2
			assert.Equal(t, "", p.color)
			assert.NoError(t, setter.Set(obj, "color", color, false))
			assert.Equal(t, color, p.color)
		})
		t.Run("***struct{}", func(t *testing.T) {
			const color = "yellow"
			p := struct {
				color string
			}{}
			p2 := &p
			p3 := &p2
			var obj any = &p3
			assert.Equal(t, "", p.color)
			assert.NoError(t, setter.Set(obj, "color", color, false))
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
			var obj any = &p4
			assert.Equal(t, "", p.color)
			assert.NoError(t, setter.Set(obj, "color", color, false))
			assert.Equal(t, color, p.color)
		})
	})
	t.Run("struct", func(t *testing.T) {
		p := person{}
		assert.NoError(t, setter.Set(&p, "Name", "Jane", false))
		assert.NoError(t, setter.Set(&p, "age", 30, true))
		assert.Equal(t, person{Name: "Jane", age: 30}, p)
	})
	t.Run("*struct", func(t *testing.T) {
		p := &person{}
		assert.NoError(t, setter.Set(&p, "Name", "Mary", false))
		assert.NoError(t, setter.Set(&p, "age", uint(33), true))
		assert.Equal(t, &person{Name: "Mary", age: 33}, p)
	})
	t.Run("var a any = &struct{}", func(t *testing.T) {
		var p any = &person{}
		assert.NoError(t, setter.Set(&p, "Name", "Mary Jane", false))
		assert.NoError(t, setter.Set(&p, "age", 45, true))
		assert.Equal(t, &person{Name: "Mary Jane", age: 45}, p)
	})
	t.Run("var a any = struct{}", func(t *testing.T) {
		var p any = person{}
		assert.NoError(t, setter.Set(&p, "Name", "Jane", false))
		assert.Equal(t, person{Name: "Jane"}, p)
	})
	t.Run("var a any = struct{}; a2 := &a; setter.Set(&a2...", func(t *testing.T) {
		var p any = person{}
		p2 := &p
		assert.NoError(t, setter.Set(&p2, "Name", "Jane", false))
		assert.Equal(t, person{Name: "Jane"}, p)
	})
	t.Run("var a1 any = struct{}; var a2 any = &a1; var a3 any = &a2; ...; setter.Set(&aN...", func(t *testing.T) {
		var p any = person{}
		p2 := &p
		var p3 any = &p2
		var p4 any = &p3
		var p5 any = &p4
		assert.NoError(t, setter.Set(&p5, "Name", "Jane", false))
		assert.Equal(t, person{Name: "Jane"}, p)
	})
	t.Run("loop #1", func(t *testing.T) {
		var p any
		p = &p
		assert.EqualError(
			t,
			setter.Set(&p, "Name", "Jane", false),
			`set (*interface {})."Name": unexpected pointer loop`,
		)
	})
	t.Run("loop #2", func(t *testing.T) {
		var a, b any
		a = &b
		b = &a
		assert.EqualError(
			t,
			setter.Set(a, "Name", "Jane", false),
			`set (*interface {})."Name": unexpected pointer loop`,
		)
	})
	t.Run("unexported type of field", func(t *testing.T) {
		p := person{}
		assert.NoError(t, setter.Set(&p, "wallet", wallet{amount: 400}, false))
		assert.Equal(t, wallet{amount: 400}, p.wallet)
	})
	t.Run("convert []any to []type", func(t *testing.T) {
		s := storage{}
		assert.NoError(
			t,
			setter.Set(&s, "wallets", []any{wallet{100}, wallet{200}}, true),
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
		err := setter.Set(&p, "Age", int16(20), true)
		assert.NoError(t, err)
		assert.Equal(t, uint(20), p.Age)
	})
	t.Run("Given errors", func(t *testing.T) {
		t.Run("Field does not exist", func(t *testing.T) {
			p := person{}
			err := setter.Set(&p, "FirstName", "Mary", false)
			assert.EqualError(t, err, `set (*setter_test.person)."FirstName": field "FirstName" does not exist`)
		})
		t.Run("Invalid pointer dest", func(t *testing.T) {
			p := 5
			err := setter.Set(&p, "FirstName", "Mary", false)
			assert.EqualError(t, err, `set (*int)."FirstName": expected pointer to struct, *int given`)
		})
		t.Run("Invalid type of value", func(t *testing.T) {
			t.Run("Convert", func(t *testing.T) {
				p := person{}
				err := setter.Set(&p, "Name", struct{}{}, true)
				assert.EqualError(t, err, `set (*setter_test.person)."Name": cannot convert struct {} to string`)
			})
			t.Run("Do not convert", func(t *testing.T) {
				p := person{}
				err := setter.Set(&p, "Name", struct{}{}, false)
				assert.EqualError(t, err, `set (*setter_test.person)."Name": value of type struct {} is not assignable to type string`)
			})
		})
		t.Run("Invalid type of value (var p any = person{})", func(t *testing.T) {
			var p any = person{}
			err := setter.Set(&p, "Name", struct{}{}, true)
			assert.EqualError(t, err, `set (*interface {})."Name": cannot convert struct {} to string`)
		})
	})
	t.Run("Invalid struct", func(t *testing.T) {
		err := setter.Set(nil, "Name", "Jane", true)
		assert.EqualError(t, err, `set (<nil>)."Name": expected pointer to struct, <nil> given`)
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
