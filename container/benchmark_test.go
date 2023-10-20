package container_test

import (
	"context"
	"testing"

	"github.com/gontainer/gontainer-helpers/container"
)

type Employee struct {
	Name string
}

func BenchmarkNewContainer_container_scopeDefault(b *testing.B) {
	c := container.NewContainer()
	e := container.NewService()
	e.SetConstructor(func() any {
		return Employee{}
	})
	e.SetField("Name", container.NewDependencyValue("Mary"))
	e.ScopeDefault()
	c.OverrideService("employee", e)
	_, _ = c.Get("employee") // warm up
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.Get("employee")
	}
}

func BenchmarkNewContainer_container_scopeShared(b *testing.B) {
	c := container.NewContainer()
	e := container.NewService()
	e.SetConstructor(func() any {
		return Employee{}
	})
	e.SetField("Name", container.NewDependencyValue("Mary"))
	e.ScopeShared()
	c.OverrideService("employee", e)
	_, _ = c.Get("employee") // warm up
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.Get("employee")
	}
}

func BenchmarkNewContainer_container_scopeContextual(b *testing.B) {
	c := container.NewContainer()
	e := container.NewService()
	e.SetConstructor(func() any {
		return Employee{}
	})
	e.SetField("Name", container.NewDependencyValue("Mary"))
	e.ScopeContextual()
	c.OverrideService("employee", e)
	_, _ = c.Get("employee") // warm up
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.Get("employee")
	}
}

func BenchmarkNewContainer_container_scopeContextual_in_same_context(b *testing.B) {
	c := container.NewContainer()
	e := container.NewService()
	e.SetConstructor(func() any {
		return Employee{}
	})
	e.SetField("Name", container.NewDependencyValue("Mary"))
	e.ScopeContextual()
	c.OverrideService("employee", e)
	ctx := container.ContextWithContainer(context.Background(), c)
	_, _ = c.GetInContext(ctx, "employee") // warm up
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.GetInContext(ctx, "employee")
	}
}

func BenchmarkNewContainer_container_scopeNonShared(b *testing.B) {
	c := container.NewContainer()
	e := container.NewService()
	e.SetConstructor(func() any {
		return Employee{}
	})
	e.SetField("Name", container.NewDependencyValue("Mary"))
	e.ScopeNonShared()
	c.OverrideService("employee", e)
	_, _ = c.Get("employee") // warm up
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.Get("employee")
	}
}

func BenchmarkNewContainer_map(b *testing.B) {
	m := make(map[string]any)
	m["employee"] = Employee{
		Name: "Mary",
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = m["employee"] //nolint:gosimple
	}
}
