package container_test

import (
	"context"
	"testing"
	"time"

	"github.com/gontainer/gontainer-helpers/v2/container"
	"github.com/stretchr/testify/require"
)

type Employee struct {
	Name string
}

func BenchmarkContainer_scopeDefault(b *testing.B) {
	c := container.New()
	e := container.NewService()
	e.SetConstructor(func() interface{} {
		return Employee{}
	})
	e.SetField("Name", container.NewDependencyParam("name"))
	e.SetScopeDefault()
	c.OverrideService("employee", e)
	c.OverrideParam("name", container.NewDependencyValue("Mary"))
	emp, _ := c.Get("employee") // warm up
	require.Equal(b, Employee{Name: "Mary"}, emp)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		time.Sleep(time.Millisecond)
		_, _ = c.Get("employee")
	}
}

func BenchmarkContainer_scopeShared(b *testing.B) {
	c := container.New()
	e := container.NewService()
	e.SetConstructor(func() interface{} {
		return Employee{}
	})
	e.SetField("Name", container.NewDependencyParam("name"))
	e.SetScopeShared()
	c.OverrideService("employee", e)
	c.OverrideParam("name", container.NewDependencyValue("Mary"))
	emp, _ := c.Get("employee") // warm up
	require.Equal(b, Employee{Name: "Mary"}, emp)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.Get("employee")
	}
}

func BenchmarkContainer_scopeContextual(b *testing.B) {
	c := container.New()
	e := container.NewService()
	e.SetConstructor(func() interface{} {
		return Employee{}
	})
	e.SetField("Name", container.NewDependencyParam("name"))
	e.SetScopeContextual()
	c.OverrideService("employee", e)
	c.OverrideParam("name", container.NewDependencyValue("Mary"))
	emp, _ := c.Get("employee") // warm up
	require.Equal(b, Employee{Name: "Mary"}, emp)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.Get("employee")
	}
}

func BenchmarkContainer_scopeContextual_in_same_context(b *testing.B) {
	c := container.New()
	e := container.NewService()
	e.SetConstructor(func() interface{} {
		return Employee{}
	})
	e.SetField("Name", container.NewDependencyParam("name"))
	e.SetScopeContextual()
	c.OverrideService("employee", e)
	c.OverrideParam("name", container.NewDependencyValue("Mary"))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = container.ContextWithContainer(ctx, c)
	emp, _ := c.GetInContext(ctx, "employee") // warm up
	require.Equal(b, Employee{Name: "Mary"}, emp)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.GetInContext(ctx, "employee")
	}
}

func BenchmarkContainer_scopeNonShared(b *testing.B) {
	c := container.New()
	e := container.NewService()
	e.SetConstructor(func() interface{} {
		return Employee{}
	})
	e.SetField("Name", container.NewDependencyParam("name"))
	e.SetScopeNonShared()
	c.OverrideService("employee", e)
	c.OverrideParam("name", container.NewDependencyValue("Mary"))
	emp, _ := c.Get("employee") // warm up
	require.Equal(b, Employee{Name: "Mary"}, emp)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.Get("employee")
	}
}

func BenchmarkContainer_map(b *testing.B) {
	m := make(map[string]interface{})
	m["employee"] = Employee{
		Name: "Mary",
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = m["employee"] //nolint:gosimple
	}
}
