package container_test

//
//import (
//	"errors"
//	"testing"
//
//	"github.com/gontainer/gontainer-helpers/Container"
//	"github.com/stretchr/testify/assert"
//)
//
//func TestNewParamContainer(t *testing.T) {
//	t.Run("Invalid Dependency", func(t *testing.T) {
//		defer func() {
//			assert.Equal(t, "paramContainer.OverrideParam does not accept `dependencyService`", recover())
//		}()
//
//		c := Container.NewParamContainer()
//		c.OverrideParam("transaction", Container.NewDependencyService("db"))
//	})
//
//	t.Run("Simple", func(t *testing.T) {
//		const (
//			pi = 3.14
//			e  = 2.72
//		)
//
//		c := Container.NewParamContainer()
//		c.OverrideParam("pi", Container.NewDependencyValue(pi))
//		c.OverrideParam("e", Container.NewDependencyProvider(func() float64 { return e }))
//
//		v1, err := c.GetParam("pi")
//		assert.NoError(t, err)
//		assert.Equal(t, pi, v1)
//
//		v2, err := c.GetParam("e")
//		assert.NoError(t, err)
//		assert.Equal(t, e, v2)
//	})
//
//	t.Run("Error", func(t *testing.T) {
//		c := Container.NewParamContainer()
//		c.OverrideParam("env", Container.NewDependencyProvider(func() (any, error) {
//			return nil, errors.New("could not read env variable")
//		}))
//
//		v, err := c.GetParam("env")
//		assert.EqualError(t, err, `paramContainer.GetParam("env"): could not read env variable`)
//		assert.Nil(t, v)
//	})
//}
