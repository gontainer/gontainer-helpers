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
	"context"
	"testing"

	"github.com/gontainer/gontainer-helpers/v2/container"
	"github.com/stretchr/testify/require"
)

type Employee struct {
	Name string
}

func (e *Employee) SetName(n string) {
	e.Name = n
}

func BenchmarkContainer_scopeDefault(b *testing.B) {
	c := container.New()
	e := container.NewService()
	e.SetConstructor(func() interface{} {
		return Employee{}
	})
	e.AppendCall("SetName", container.NewDependencyParam("name"))
	e.SetScopeDefault()
	c.OverrideService("employee", e)
	c.OverrideParam("name", container.NewDependencyValue("Mary"))
	emp, _ := c.Get("employee") // warm up
	require.Equal(b, Employee{Name: "Mary"}, emp)
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
	e.AppendCall("SetName", container.NewDependencyParam("name"))
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
	e.AppendCall("SetName", container.NewDependencyParam("name"))
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

func BenchmarkContainer_scopeContextualSetField(b *testing.B) {
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
	e.AppendCall("SetName", container.NewDependencyParam("name"))
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
	e.AppendCall("SetName", container.NewDependencyParam("name"))
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
