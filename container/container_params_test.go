package container_test

import (
	"errors"
	"testing"

	"github.com/gontainer/gontainer-helpers/container"
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
}
