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

package accessors_test

import (
	"testing"

	"github.com/gontainer/gontainer-helpers/v2/accessors"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	type Person struct {
		name string
	}

	t.Run("OK", func(t *testing.T) {
		t.Run("#1", func(t *testing.T) {
			var p any = Person{name: "Mary"}
			n, err := accessors.Get(p, "name")
			assert.NoError(t, err)
			assert.Equal(t, "Mary", n)
		})
		t.Run("#2", func(t *testing.T) {
			var p any = &Person{name: "Mary"}
			n, err := accessors.Get(p, "name")
			assert.NoError(t, err)
			assert.Equal(t, "Mary", n)
		})
		t.Run("#3", func(t *testing.T) {
			var p any = &Person{name: "Mary"}
			n, err := accessors.Get(&p, "name")
			assert.NoError(t, err)
			assert.Equal(t, "Mary", n)
		})
		t.Run("#4", func(t *testing.T) {
			p := Person{name: "Mary"}
			n, err := accessors.Get(p, "name")
			assert.NoError(t, err)
			assert.Equal(t, "Mary", n)
		})
		t.Run("#5", func(t *testing.T) {
			p := &Person{name: "Mary"}
			n, err := accessors.Get(&p, "name")
			assert.NoError(t, err)
			assert.Equal(t, "Mary", n)
		})
		t.Run("#5", func(t *testing.T) {
			p := &Person{name: "Mary"}
			var p2 any = &p
			n, err := accessors.Get(&p2, "name")
			assert.NoError(t, err)
			assert.Equal(t, "Mary", n)
		})
	})

	t.Run("Errors", func(t *testing.T) {
		t.Run("#1", func(t *testing.T) {
			_, err := accessors.Get(nil, "name")
			assert.EqualError(t, err, `get (<nil>)."name": expected struct, <nil> given`)
		})
		t.Run("#2", func(t *testing.T) {
			_, err := accessors.Get(make([]int, 0), "name")
			assert.EqualError(t, err, `get ([]int)."name": expected struct, []int given`)
		})
		t.Run("#3", func(t *testing.T) {
			_, err := accessors.Get(Person{}, "_")
			assert.EqualError(t, err, `get (accessors_test.Person)."_": "_" is not supported`)
		})
		t.Run("#4", func(t *testing.T) {
			_, err := accessors.Get(Person{name: "Mary"}, "age")
			assert.EqualError(t, err, `get (accessors_test.Person)."age": field "age" does not exist`)
		})
		t.Run("#5", func(t *testing.T) {
			var p any
			p = &p
			_, err := accessors.Get(p, "age")
			assert.EqualError(t, err, `get (*interface {})."age": unexpected pointer loop`)
		})
	})
}
