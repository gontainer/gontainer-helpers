package container_test

import (
	"testing"

	"github.com/gontainer/gontainer-helpers/container"
)

type Employee struct {
	Name string
}

func BenchmarkNewContainer_container(b *testing.B) {
	c := container.NewContainer()
	e := container.NewService()
	e.SetConstructor(func() interface{} {
		return Employee{}
	})
	e.SetField("Name", container.NewDependencyValue("Mary"))
	c.OverrideService("employee", e)
	_, _ = c.Get("employee")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.Get("employee")
	}
}

func BenchmarkNewContainer_map(b *testing.B) {
	m := make(map[string]interface{})
	m["employee"] = Employee{
		Name: "Mary",
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = m["employee"]
	}
}
