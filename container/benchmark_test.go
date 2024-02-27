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

package container_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gontainer/gontainer-helpers/v3/container"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/dependency"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/service"
	"github.com/stretchr/testify/require"
)

type Employee struct {
	Name string
}

func NewEmployee(n string) Employee {
	return Employee{Name: n}
}

func (e *Employee) SetName(n string) {
	e.Name = n
}

func (e Employee) WithName(n string) Employee {
	e.Name = n
	return e
}

func BenchmarkContainer_scopeDefault(b *testing.B) {
	c := container.New()
	e := service.New()
	e.
		SetConstructor(func() interface{} {
			return Employee{}
		}).
		AppendCall("SetName", dependency.Param("name")).
		SetScopeDefault()
	c.OverrideService("employee", e)
	c.OverrideParam("name", dependency.Value("Mary"))
	emp, _ := c.Get("employee") // warm up
	require.Equal(b, Employee{Name: "Mary"}, emp)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.Get("employee")
	}
}

func BenchmarkContainer_scopeShared(b *testing.B) {
	c := container.New()
	e := service.New()
	e.
		SetConstructor(func() interface{} {
			return Employee{}
		}).
		AppendCall("SetName", dependency.Param("name")).
		SetScopeShared()
	c.OverrideService("employee", e)
	c.OverrideParam("name", dependency.Value("Mary"))
	emp, _ := c.Get("employee") // warm up
	require.Equal(b, Employee{Name: "Mary"}, emp)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.Get("employee")
	}
}

func BenchmarkContainer_scopeContextual_useConstructor(b *testing.B) {
	c := container.New()
	e := service.New()
	e.
		SetConstructor(
			NewEmployee,
			dependency.Param("name"),
		).
		SetScopeContextual()
	c.OverrideService("employee", e)
	c.OverrideParam("name", dependency.Value("Mary"))
	emp, _ := c.Get("employee") // warm up
	require.Equal(b, Employee{Name: "Mary"}, emp)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.Get("employee")
	}
}

func BenchmarkContainer_scopeContextual_AppendCall(b *testing.B) {
	c := container.New()
	e := service.New()
	e.
		SetConstructor(func() interface{} {
			return Employee{}
		}).
		AppendCall("SetName", dependency.Param("name")).
		SetScopeContextual()
	c.OverrideService("employee", e)
	c.OverrideParam("name", dependency.Value("Mary"))
	emp, _ := c.Get("employee") // warm up
	require.Equal(b, Employee{Name: "Mary"}, emp)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.Get("employee")
	}
}

func BenchmarkContainer_scopeContextual_SetField(b *testing.B) {
	c := container.New()
	e := service.New()
	e.
		SetConstructor(func() interface{} {
			return Employee{}
		}).
		SetField("Name", dependency.Param("name")).
		SetScopeContextual()
	c.OverrideService("employee", e)
	c.OverrideParam("name", dependency.Value("Mary"))
	emp, _ := c.Get("employee") // warm up
	require.Equal(b, Employee{Name: "Mary"}, emp)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.Get("employee")
	}
}

func BenchmarkContainer_scopeContextual_AppendWither(b *testing.B) {
	c := container.New()
	e := service.New()
	e.
		SetConstructor(func() interface{} {
			return Employee{}
		}).
		AppendWither("WithName", dependency.Param("name")).
		SetScopeContextual()
	c.OverrideService("employee", e)
	c.OverrideParam("name", dependency.Value("Mary"))
	emp, _ := c.Get("employee") // warm up
	require.Equal(b, Employee{Name: "Mary"}, emp)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.Get("employee")
	}
}

func BenchmarkContainer_scopeContextual_in_same_context(b *testing.B) {
	c := container.New()
	e := service.New()
	e.
		SetConstructor(func() interface{} {
			return Employee{}
		}).
		AppendCall("SetName", dependency.Param("name")).
		SetScopeContextual()
	c.OverrideService("employee", e)
	c.OverrideParam("name", dependency.Value("Mary"))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = container.ContextWithContainer(ctx, c)
	emp, _ := c.GetInContext(ctx, "employee") // warm up
	require.Equal(b, Employee{Name: "Mary"}, emp)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.GetInContext(ctx, "employee")
	}
}

func BenchmarkContainer_scopeNonShared_AppendCall(b *testing.B) {
	c := container.New()
	e := service.New()
	e.
		SetConstructor(func() interface{} {
			return Employee{}
		}).
		AppendCall("SetName", dependency.Param("name")).
		SetScopeNonShared()
	c.OverrideService("employee", e)
	c.OverrideParam("name", dependency.Value("Mary"))
	emp, _ := c.Get("employee") // warm up
	require.Equal(b, Employee{Name: "Mary"}, emp)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.Get("employee")
	}
}

type serviceMap map[string]interface{}

func (s serviceMap) Get(id string) (interface{}, error) {
	v, ok := s[id]
	if !ok {
		return nil, errors.New("does not exist")
	}
	return v, nil
}

func (s serviceMap) Set(id string, v interface{}) {
	s[id] = v
}

func BenchmarkContainer_map(b *testing.B) {
	m := make(serviceMap)
	m.Set("employee", Employee{
		Name: "Mary",
	})
	emp, _ := m.Get("employee")
	require.Equal(b, Employee{Name: "Mary"}, emp)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = m.Get("employee")
	}
}
