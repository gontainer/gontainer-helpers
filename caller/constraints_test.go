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
	"testing"

	"github.com/gontainer/gontainer-helpers/v2/caller"
	"github.com/stretchr/testify/assert"
)

type book struct {
	title string
}

func (b *book) SetTitle(t string) {
	b.title = t
}

func (b *book) setTitle(t string) { //nolint:unused
	b.title = t
}

func (b book) WithTitle(t string) book {
	b.title = t
	return b
}

func TestConstraint(t *testing.T) {
	const (
		harryPotterTitle = "Harry Potter"
	)

	var (
		harryPotter = book{title: harryPotterTitle}
		emptyBook   = book{}
	)

	// https://github.com/golang/go/wiki/MethodSets#interfaces

	// Method with a pointer receiver requires explicit definition of the pointer:
	// v := &book{}; CallByName(v, ...
	// var v any = &book{}; CallByName(v, ...
	// v := book{}; CallByName(&v, ...
	//
	// Creating variable as a value will not work:
	// v := book{}; CallByName(v, ...
	// var v interface{} = book{}; CallByName(&v, ...
	t.Run("Call a method", func(t *testing.T) {
		t.Run("A pointer receiver", func(t *testing.T) {
			t.Run("Given errors", func(t *testing.T) {
				t.Run("v := book{}; CallByName(v, ...", func(t *testing.T) {
					b := book{}
					r, err := caller.CallByName(b, "SetTitle", []any{harryPotterTitle}, false)
					assert.EqualError(t, err, `cannot call method (caller_test.book)."SetTitle": invalid func (caller_test.book)."SetTitle"`)
					assert.Nil(t, r)
					assert.Zero(t, b)
				})
				t.Run("var v any = book{}; CallByName(&v, ...", func(t *testing.T) {
					var b any = book{}
					r, err := caller.CallByName(&b, "SetTitle", []any{harryPotterTitle}, false)
					assert.EqualError(t, err, `cannot call method (*interface {})."SetTitle": invalid func (*interface {})."SetTitle"`)
					assert.Nil(t, r)
					assert.Equal(t, emptyBook, b)
				})
			})
			t.Run("Given scenarios", func(t *testing.T) {
				t.Run("v := book{}; CallByName(&v, ...", func(t *testing.T) {
					b := book{}
					r, err := caller.CallByName(&b, "SetTitle", []any{harryPotterTitle}, false)
					assert.NoError(t, err)
					assert.Len(t, r, 0)
					assert.Equal(t, harryPotter, b)
				})
				t.Run("v := &book{}; CallByName(&v, ...", func(t *testing.T) {
					b := &book{}
					r, err := caller.CallByName(&b, "SetTitle", []any{harryPotterTitle}, false)
					assert.NoError(t, err)
					assert.Len(t, r, 0)
					assert.Equal(t, &harryPotter, b)
				})
				t.Run("v := &book{}; CallByName(v, ...", func(t *testing.T) {
					b := &book{}
					r, err := caller.CallByName(b, "SetTitle", []any{harryPotterTitle}, false)
					assert.NoError(t, err)
					assert.Len(t, r, 0)
					assert.Equal(t, &harryPotter, b)
				})
				t.Run("var v any = &book{}; CallByName(v, ...", func(t *testing.T) {
					var b any = &book{}
					r, err := caller.CallByName(b, "SetTitle", []any{harryPotterTitle}, false)
					assert.NoError(t, err)
					assert.Len(t, r, 0)
					assert.Equal(t, &harryPotter, b)
				})
				t.Run("var v any = &book{}; CallByName(&v, ...", func(t *testing.T) {
					var b any = &book{}
					r, err := caller.CallByName(&b, "SetTitle", []any{harryPotterTitle}, false)
					assert.NoError(t, err)
					assert.Len(t, r, 0)
					assert.Equal(t, &harryPotter, b)
				})
				t.Run("var v interface{ SetTitle(string) } = &book{}; CallByName(v, ...", func(t *testing.T) {
					var b interface{ SetTitle(string) } = &book{}
					r, err := caller.CallByName(b, "SetTitle", []any{harryPotterTitle}, false)
					assert.NoError(t, err)
					assert.Len(t, r, 0)
					assert.Equal(t, &harryPotter, b)
				})
			})
		})
		// Methods with a value receiver do not have any constraints
		t.Run("A value receiver", func(t *testing.T) {
			t.Run("b := book{}", func(t *testing.T) {
				b := book{}
				r, err := caller.CallWitherByName(b, "WithTitle", []any{harryPotterTitle}, false)
				assert.NoError(t, err)
				assert.Equal(t, harryPotter, r)
				assert.Zero(t, b)
			})
			t.Run("b := &book{}", func(t *testing.T) {
				b := &book{}
				r, err := caller.CallWitherByName(b, "WithTitle", []any{harryPotterTitle}, false)
				assert.NoError(t, err)
				assert.Equal(t, harryPotter, r)
				assert.Equal(t, &emptyBook, b)
			})
			t.Run("var b any = book{}", func(t *testing.T) {
				var b any = book{}
				r, err := caller.CallWitherByName(b, "WithTitle", []any{harryPotterTitle}, false)
				assert.NoError(t, err)
				assert.Equal(t, harryPotter, r)
				assert.Equal(t, emptyBook, b)
			})
			t.Run("var b any = &book{}", func(t *testing.T) {
				var b any = &book{}
				r, err := caller.CallWitherByName(b, "WithTitle", []any{harryPotterTitle}, false)
				assert.NoError(t, err)
				assert.Equal(t, harryPotter, r)
				assert.Equal(t, &emptyBook, b)
			})
		})
		t.Run("An unexported method", func(t *testing.T) {
			b := book{}
			_, err := caller.CallByName(&b, "setTitle", []any{harryPotter}, false)
			assert.EqualError(t, err, `cannot call method (*caller_test.book)."setTitle": invalid func (*caller_test.book)."setTitle"`)
		})
	})
}
