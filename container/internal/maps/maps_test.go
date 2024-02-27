// Copyright (c) 2023–present Bartłomiej Krukowski
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

package maps_test

import (
	"fmt"
	"testing"

	"github.com/gontainer/gontainer-helpers/v3/container/internal/maps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSortedStringKeys(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		t.Run("#1", func(t *testing.T) {
			input := map[string]struct{}{
				"orange": {},
				"apple":  {},
				"banana": {},
			}
			expected := []string{"apple", "banana", "orange"}

			for i := 0; i < 100; i++ {
				require.Equal(t, expected, maps.SortedStringKeys(input))
			}
		})
		t.Run("#2", func(t *testing.T) {
			type myMap map[string]struct{}
			input := myMap{
				"orange": {},
				"apple":  {},
				"banana": {},
			}
			expected := []string{"apple", "banana", "orange"}

			for i := 0; i < 100; i++ {
				require.Equal(t, expected, maps.SortedStringKeys(input))
			}
		})
	})
	t.Run("Nil", func(t *testing.T) {
		require.Nil(t, maps.SortedStringKeys((map[string]bool)(nil)))
	})
	t.Run("Panic", func(t *testing.T) {
		type myString string
		type myMap map[myString]struct{}

		scenarios := []struct {
			input any
			panic string
		}{
			{
				input: myMap{},
				panic: "SortedStringKeys: expected map[string]T, maps_test.myMap given",
			},
			{
				input: nil,
				panic: "SortedStringKeys: expected map[string]T, <nil> given",
			},
		}

		for i, s := range scenarios {
			s := s
			t.Run(fmt.Sprintf("Scenario #%d", i), func(t *testing.T) {
				defer func() {
					assert.Equal(t, s.panic, recover())
				}()
				maps.SortedStringKeys(s.input)
			})
		}
	})
}
