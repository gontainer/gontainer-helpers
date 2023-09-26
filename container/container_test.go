package container_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gontainer/gontainer-helpers/container"
	errAssert "github.com/gontainer/gontainer-helpers/errors/assert"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Numbers struct {
	A, B, C, D int64
}

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

func Test_container_Get(t *testing.T) {
	t.Run("ContextualBag & concurrency", func(t *testing.T) {
		c := container.NewContainer()

		next := int64(0)
		svcNextInt := container.NewService()
		svcNextInt.SetConstructor(func() int64 {
			return atomic.AddInt64(&next, 1)
		})
		svcNextInt.ScopeContextual()
		c.OverrideService("nextInt", svcNextInt)

		svcNum := container.NewService()
		svcNum.
			SetConstructor(func() *Numbers { return &Numbers{} }).
			SetField("A", container.NewDependencyService("nextInt")).
			SetField("B", container.NewDependencyService("nextInt")).
			SetField("C", container.NewDependencyService("nextInt")).
			SetField("D", container.NewDependencyService("nextInt"))

		const max = 5000

		for i := 0; i < max; i++ {
			c.OverrideService(fmt.Sprintf("numbers%d", i), svcNum)
		}

		wg := sync.WaitGroup{}
		for i := 0; i < max; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				tmp, err := c.Get(fmt.Sprintf("numbers%d", i))
				assert.NoError(t, err)

				nums := tmp.(*Numbers)
				assert.Equal(t, nums.A, nums.B)
				assert.Equal(t, nums.A, nums.C)
				assert.Equal(t, nums.A, nums.D)

			}(i)
		}
		wg.Wait()

		tmp, err := c.Get("numbers0")
		require.NoError(t, err)
		nums := tmp.(*Numbers)
		assert.Equal(t, nums.A, int64(max+1))
		assert.Equal(t, nums.B, int64(max+1))
		assert.Equal(t, nums.C, int64(max+1))
		assert.Equal(t, nums.D, int64(max+1))
	})

	t.Run("Circular dependencies", func(t *testing.T) {
		c := container.NewContainer()

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

		svc, err := c.Get("s1")
		assert.Nil(t, svc)

		expected := []string{
			`container.get("s1"): circular dependencies: @s1 -> @s2 -> @s3 -> @s1`,
			`container.get("s1"): circular dependencies: @s1 -> @s1`,
		}
		errAssert.EqualErrorGroup(t, err, expected)
	})

	t.Run("ContextualScope", func(t *testing.T) {
		t.Run("Default scope", func(t *testing.T) {
			c := container.NewContainer()

			transaction := container.NewService()
			transaction.SetConstructor(func() Transaction {
				var t int
				return &t
			})
			transaction.ScopeContextual()
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
				var svc MyService
				assert.NoError(t, c.CopyServiceTo("myService", &svc))

				assert.Same(t, svc.Transaction, svc.ItemStorage.Transaction)
				assert.Same(t, svc.Transaction, svc.UserStorage.Transaction)

				var svc2 MyService
				assert.NoError(t, c.CopyServiceTo("myService", &svc2))
				assert.NotSame(t, svc.Transaction, svc2.Transaction)
			}()

			func() {
				transaction.ScopeNonShared()
				c.OverrideService("transaction", transaction)

				var svc MyService
				assert.NoError(t, c.CopyServiceTo("myService", &svc))

				assert.NotSame(t, svc.Transaction, svc.ItemStorage.Transaction)
				assert.NotSame(t, svc.Transaction, svc.UserStorage.Transaction)
			}()
		})
	})
}

func Test_container_CopyServiceTo(t *testing.T) {
	t.Run("OK", func(t *testing.T) {
		c := container.NewContainer()

		s := container.NewService()
		s.SetConstructor(func() interface{} {
			return &Numbers{
				A: 200,
			}
		})
		c.OverrideService("numbers", s)

		var numbers *Numbers
		err := c.CopyServiceTo("numbers", &numbers)
		assert.NoError(t, err)
		assert.Equal(t, Numbers{A: 200}, *numbers)
	})
	t.Run("Error", func(t *testing.T) {
		c := container.NewContainer()

		s := container.NewService()
		s.SetConstructor(func() interface{} {
			return &Numbers{
				A: 200,
			}
		})
		c.OverrideService("numbers", s)

		var numbers *Numbers
		err := c.CopyServiceTo("numbers", numbers) // missing pointer, it should be &numbers
		assert.ErrorContains(t, err, `container.CopyServiceTo("numbers"): `)
	})
}

func Test_container_CircularDeps(t *testing.T) {
	// since we iterate over map `g.container.services` (`map[string]Service`) in the method `graphBuilder.warmUp`,
	// the order of errors can differ,
	// so we need to run these tests many times to make sure we have consistent results always
	for i := 0; i < 50; i++ {
		c := container.NewContainer()

		s1 := container.NewService()
		s1.SetField("service1", container.NewDependencyService("service1"))
		s1.SetField("service2", container.NewDependencyService("service2"))
		c.OverrideService("service1", s1)

		s2 := container.NewService()
		s2.SetField("service1", container.NewDependencyService("service1"))
		c.OverrideService("service2", s2)

		expected := []string{
			`container.CircularDeps(): @service1 -> @service1`,
			`container.CircularDeps(): @service1 -> @service2 -> @service1`,
		}
		errAssert.EqualErrorGroup(t, c.CircularDeps(), expected)
	}
}
