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

package container_test

import (
	"errors"
	"testing"

	"github.com/gontainer/gontainer-helpers/v2/container"
	"github.com/stretchr/testify/assert"
)

func TestContainer_GetParam(t *testing.T) {
	t.Run("Invalid dependency", func(t *testing.T) {
		defer func() {
			assert.Equal(t, "overrideParam: invalid dependency: dependencyService", recover())
		}()

		c := container.New()
		c.OverrideParam("transaction", container.NewDependencyService("db"))
	})

	t.Run("Simple", func(t *testing.T) {
		const (
			pi = 3.14
			e  = 2.72
		)

		c := container.New()
		c.OverrideParam("pi", container.NewDependencyValue(pi))
		c.OverrideParam("e", container.NewDependencyProvider(func() float64 { return e }))

		v1, err := c.GetParam("pi")
		assert.NoError(t, err)
		assert.Equal(t, pi, v1)

		v2, err := c.GetParam("e")
		assert.NoError(t, err)
		assert.Equal(t, e, v2)
	})

	t.Run("Error", func(t *testing.T) {
		c := container.New()
		c.OverrideParam("env", container.NewDependencyProvider(func() (any, error) {
			return nil, errors.New("could not read env variable")
		}))

		v, err := c.GetParam("env")
		assert.EqualError(t, err, `getParam("env"): cannot call provider func() (interface {}, error): could not read env variable`)
		assert.Nil(t, v)
	})

	t.Run("Param does not exist", func(t *testing.T) {
		c := container.New()

		v, err := c.GetParam("myParam")
		assert.EqualError(t, err, `getParam("myParam"): param does not exist`)
		assert.Nil(t, v)
	})
}
