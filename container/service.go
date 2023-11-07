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

package container

import (
	"fmt"
	"reflect"
	"sort"
)

type serviceCall struct {
	wither bool
	method string
	deps   []Dependency
}

type serviceField struct {
	name string
	dep  Dependency
}

// Service represents a service in the [*Container].
// Use [NewService] to create a new instance.
type Service struct {
	hasCreationMethod bool
	value             any
	constructor       any
	constructorDeps   []Dependency
	factoryServiceID  string
	factoryMethod     string
	factoryDeps       []Dependency
	calls             []serviceCall
	fields            []serviceField
	tags              map[string]int
	scope             scope
}

// NewService creates a new service.
func NewService() Service {
	return Service{
		tags:  make(map[string]int),
		scope: scopeDefault,
	}
}

func (s *Service) resetCreationMethods() {
	s.value = nil
	s.constructor = nil
	s.constructorDeps = nil
	s.factoryServiceID = ""
	s.factoryMethod = ""
	s.factoryDeps = nil
}

/*
SetValue sets a predefined value of the service. It excludes [*Service.SetConstructor].
It panics for pointers, channels, maps, and slices, otherwise we could have issues with scopes.

Example of error-prone code:

	c := container.New()
	svc := container.NewService()
	svc.SetValue(make(chan struct{}))
	svc.SetScopeNonShared() // scope is non shared, so we expect a new chan for each invocation
	c.OverrideService("chan", svc)

	chan1, _ := c.Get("chan")
	chan2, _ := c.Get("chan")
	fmt.Println(chan1 == chan2) // true - it's unexpected

Using a constructor solves that problem:

	svc.SetConstructor(func() chan struct{} {
		return make(chan struct{})
	})
*/
func (s *Service) SetValue(v any) *Service {
	k := reflect.ValueOf(v).Kind()
	switch reflect.ValueOf(v).Kind() {
	case
		reflect.Ptr,
		reflect.Chan,
		reflect.Map,
		reflect.Slice:
		panic(fmt.Sprintf("container.Service: passing %s to SetValue is error-prone, use SetConstructor instead", k))
	}

	s.resetCreationMethods()
	s.value = v
	s.hasCreationMethod = true
	return s
}

/*
SetConstructor sets a constructor of the service. It excludes [*Service.SetValue].

	s := container.NewService()
	s.SetConstructor(
		func NewDB(username, password string) (*sql.DB, error) {
			return sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/test", username, password))
		},
		container.NewDependencyValue("root"),
		container.NewDependencyValue("root"),
	)
*/
func (s *Service) SetConstructor(fn any, deps ...Dependency) *Service {
	s.resetCreationMethods()
	s.constructor = fn
	s.constructorDeps = deps
	s.hasCreationMethod = true
	return s
}

/*
SetFactory sets a factory that is supposed to return the given service.

	// "db" refers to *sql.DB
	s := container.NewService()
	s.SetFactory("db", "BeginTx", container.NewDependencyContext(), container.NewDependencyValue(nil))
	s.SetScopeContextual()
*/
func (s *Service) SetFactory(serviceID string, method string, deps ...Dependency) *Service {
	if serviceID == "" {
		panic(`serviceID == ""`)
	}
	if method == "" {
		panic(`method == ""`)
	}

	s.resetCreationMethods()
	s.factoryServiceID = serviceID
	s.factoryMethod = method
	s.factoryDeps = deps
	s.hasCreationMethod = true
	return s
}

/*
AppendCall instructs the container to execute a method over that object.

	s := container.NewService()
	s.SetConstructor(func() *Person {
		return &Person{}
	})
	s.AppendMethod("SetName", container.NewDependencyValue("Jane"))

	// p := &Person{}
	// p.SetName("Jane")
*/
func (s *Service) AppendCall(method string, deps ...Dependency) *Service {
	s.calls = append(s.calls, serviceCall{
		wither: false,
		method: method,
		deps:   deps,
	})
	return s
}

/*
AppendWither instructs the container to execute a wither-method over that object.

	s := container.NewService()
	s.SetValue(Person{})
	s.AppendWither("WithName", container.NewDependencyValue("Jane"))

	// p := Person{}
	// p = p.WithName("Jane")
*/
func (s *Service) AppendWither(method string, deps ...Dependency) *Service {
	s.calls = append(s.calls, serviceCall{
		wither: true,
		method: method,
		deps:   deps,
	})
	return s
}

/*
SetField instructs the container to set a value of the given field.

	s := container.NewService()
	s.SetValue(Person{})
	s.SetField("Name", container.NewDependencyValue("Jane"))

	// p := Person{}
	// p.Name = "Jane"
*/
func (s *Service) SetField(field string, dep Dependency) *Service {
	s.fields = append(s.fields, serviceField{
		name: field,
		dep:  dep,
	})

	// check if we have a duplicate, if yes, remove it
	for i := 0; i < len(s.fields)-1; i++ {
		if s.fields[i].name == field {
			s.fields = append(s.fields[:i], s.fields[i+1:]...)
			break
		}
	}
	return s
}

/*
SetFields instructs the container to set many fields to the given struct.

	s := container.NewService()
	s.SetValue(struct {
		Name string
		Age  int
	}{})
	s.SetFields(map[string]container.Dependency{
		"Name": container.NewDependencyValue("Jane"),
		"Age":  container.NewDependencyValue(30),
	})

See [*Service.SetField].
*/
func (s *Service) SetFields(fields map[string]Dependency) *Service {
	// sort names to have the same order of errors always
	names := make([]string, 0, len(fields))
	for n := range fields {
		names = append(names, n)
	}
	sort.Strings(names)

	for _, n := range names {
		s.SetField(n, fields[n])
	}
	return s
}

// Tag tags the given service. Argument priority it is being used for determining order in [*Container.GetTaggedBy].
func (s *Service) Tag(tag string, priority int) *Service {
	s.tags[tag] = priority
	return s
}

// SetScopeDefault sets the scope that will be determined in real-time.
func (s *Service) SetScopeDefault() *Service {
	s.scope = scopeDefault
	return s
}

// SetScopeShared sets the shared scope.
func (s *Service) SetScopeShared() *Service {
	s.scope = scopeShared
	return s
}

// SetScopeContextual sets the contextual scope.
//
// See [*Container.GetInContext].
func (s *Service) SetScopeContextual() *Service {
	s.scope = scopeContextual
	return s
}

// SetScopeNonShared sets the non-shared scope.
// Each invocation of [*Container.Get] will create a new variable.
func (s *Service) SetScopeNonShared() *Service {
	s.scope = scopeNonShared
	return s
}
