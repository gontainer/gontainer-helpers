package container_test

import (
	"context"
	"testing"

	"github.com/gontainer/gontainer-helpers/v2/container"
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
	_, _ = c.Get("employee") // warm up
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
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
	_, _ = c.Get("employee") // warm up
	b.ReportAllocs()
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
	_, _ = c.Get("employee") // warm up
	b.ReportAllocs()
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
	_, _ = c.GetInContext(ctx, "employee") // warm up
	b.ReportAllocs()
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
	_, _ = c.Get("employee") // warm up
	b.ReportAllocs()
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
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = m["employee"] //nolint:gosimple
	}
}
