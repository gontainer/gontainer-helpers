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

package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_SetField(t *testing.T) {
	t.Run("Remove duplicates", func(t *testing.T) {
		s := NewService()
		s.SetField("age", NewDependencyValue(30))
		for _, n := range []string{"Jane", "John", "Mary"} {
			s.SetField("name", NewDependencyValue(n))
		}
		s.SetField("eyeColor", NewDependencyValue("blue"))
		require.Len(t, s.fields, 3)
		assert.Equal(t, "age", s.fields[0].name)
		assert.Equal(t, "name", s.fields[1].name)
		assert.Equal(t, "Mary", s.fields[1].dep.value)
		assert.Equal(t, "eyeColor", s.fields[2].name)
	})
}

func TestService_SetFields(t *testing.T) {
	s := NewService()
	s.SetFields(map[string]Dependency{
		"lastname":  NewDependencyValue("Stark"),
		"firstname": NewDependencyValue("Tony"),
	})
	assert.Equal(
		t,
		[]serviceField{
			{
				name: "firstname",
				dep: Dependency{
					type_: dependencyValue,
					value: "Tony",
				},
			},
			{
				name: "lastname",
				dep: Dependency{
					type_: dependencyValue,
					value: "Stark",
				},
			},
		},
		s.fields,
	)
}
