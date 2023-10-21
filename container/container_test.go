package container_test

import (
	"testing"

	"github.com/gontainer/gontainer-helpers/container"
	errAssert "github.com/gontainer/gontainer-helpers/grouperror/assert"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainer_Get(t *testing.T) {
	t.Run("Circular dependencies", func(t *testing.T) {
		c := container.New()

		s1 := container.NewService()
		s1.SetField("Dependency", container.NewDependencyService("s2"))
		s1.SetField("dependency2", container.NewDependencyService("s1"))
		c.OverrideService("s1", s1)

		s2 := container.NewService()
		s2.SetField("Dependency", container.NewDependencyService("s3"))
		c.OverrideService("s2", s2)

		s3 := container.NewService()
		s3.SetField("Dependency", container.NewDependencyService("s1"))
		c.OverrideService("s3", s3)

		c.OverrideParam("lastname", container.NewDependencyParam("lastname"))
		c.OverrideParam("fullname", container.NewDependencyParam("name"))
		c.OverrideParam("name", container.NewDependencyParam("fullname"))

		svc, err := c.Get("s1")
		assert.Nil(t, svc)
		expected := []string{
			`get("s1"): circular dependencies: @s1 -> @s2 -> @s3 -> @s1`,
			`get("s1"): circular dependencies: @s1 -> @s1`,
		}
		errAssert.EqualErrorGroup(t, err, expected)

		param, err := c.GetParam("fullname")
		assert.Nil(t, param)
		expected = []string{
			`getParam("fullname"): circular dependencies: %fullname% -> %name% -> %fullname%`,
		}
		errAssert.EqualErrorGroup(t, err, expected)

		expected = []string{
			`CircularDeps(): @s1 -> @s2 -> @s3 -> @s1`,
			`CircularDeps(): @s1 -> @s1`,
			`CircularDeps(): %fullname% -> %name% -> %fullname%`,
			`CircularDeps(): %lastname% -> %lastname%`,
		}
		errAssert.EqualErrorGroup(t, c.CircularDeps(), expected)
	})

	t.Run("ContextualScope", func(t *testing.T) {
		t.Run("Default scope", func(t *testing.T) {
			type Transaction *int

			type UserStorage struct {
				Transaction Transaction
			}

			type ItemStorage struct {
				Transaction Transaction
			}

			type MyService struct {
				UserStorage UserStorage
				ItemStorage ItemStorage
				Transaction Transaction
			}

			c := container.New()

			transaction := container.NewService()
			transaction.SetConstructor(func() Transaction {
				var t int
				return &t
			})
			transaction.SetScopeContextual()
			c.OverrideService("transaction", transaction)

			userStorage := container.NewService()
			userStorage.SetValue(UserStorage{})
			userStorage.SetField("Transaction", container.NewDependencyService("transaction"))
			c.OverrideService("userStorage", userStorage)

			itemStorage := container.NewService()
			itemStorage.SetValue(ItemStorage{})
			itemStorage.SetField("Transaction", container.NewDependencyService("transaction"))
			c.OverrideService("itemStorage", itemStorage)

			myService := container.NewService()
			myService.SetValue(MyService{})
			myService.SetField("UserStorage", container.NewDependencyService("userStorage"))
			myService.SetField("ItemStorage", container.NewDependencyService("itemStorage"))
			myService.SetField("Transaction", container.NewDependencyService("transaction"))
			c.OverrideService("myService", myService)

			func() {
				tmp, err := c.Get("myService")
				require.NoError(t, err)
				svc := tmp.(MyService)

				assert.Same(t, svc.Transaction, svc.ItemStorage.Transaction)
				assert.Same(t, svc.Transaction, svc.UserStorage.Transaction)

				tmp, err = c.Get("myService")
				require.NoError(t, err)
				svc2 := tmp.(MyService)
				assert.NotSame(t, svc.Transaction, svc2.Transaction)
			}()

			func() {
				transaction.SetScopeNonShared()
				c.OverrideService("transaction", transaction)

				tmp, err := c.Get("myService")
				require.NoError(t, err)
				svc := tmp.(MyService)

				assert.NotSame(t, svc.Transaction, svc.ItemStorage.Transaction)
				assert.NotSame(t, svc.Transaction, svc.UserStorage.Transaction)
			}()
		})
	})
}

func TestContainer_CircularDeps(t *testing.T) {
	// since we iterate over maps `g.Container.services`, and `g.Container.params` in the method `graphBuilder.warmUp`,
	// the order of errors can differ,
	// so we need to run these tests many times to make sure we have consistent results always
	for i := 0; i < 50; i++ {
		c := container.New()

		s1 := container.NewService()
		s1.SetField("service1", container.NewDependencyService("service1"))
		s1.SetField("service2", container.NewDependencyService("service2"))
		c.OverrideService("service1", s1)

		s2 := container.NewService()
		s2.SetField("service1", container.NewDependencyService("service1"))
		c.OverrideService("service2", s2)

		c.OverrideParam("name", container.NewDependencyParam("name"))
		c.OverrideParam("a", container.NewDependencyParam("b"))
		c.OverrideParam("b", container.NewDependencyParam("c"))
		c.OverrideParam("c", container.NewDependencyParam("a"))

		expected := []string{
			`CircularDeps(): @service1 -> @service1`,
			`CircularDeps(): @service1 -> @service2 -> @service1`,
			`CircularDeps(): %a% -> %b% -> %c% -> %a%`,
			`CircularDeps(): %name% -> %name%`,
		}
		errAssert.EqualErrorGroup(t, c.CircularDeps(), expected)
	}
}
