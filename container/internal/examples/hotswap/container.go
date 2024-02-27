// Copyright (c) 2023-2024 Bartłomiej Krukowski
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

package hotswap

import (
	"net/http"

	"github.com/gontainer/gontainer-helpers/v3/container"
	containerHttp "github.com/gontainer/gontainer-helpers/v3/container/http"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/dependency"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/param"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/service"
)

type Container struct {
	*container.Container
}

func (c *Container) HTTPHandler() http.Handler {
	s, err := c.Get("httpHandler")
	if err != nil {
		panic(err)
	}
	return s.(http.Handler)
}

func (c *Container) ParamA() int {
	p, err := c.GetParam("a")
	if err != nil {
		panic(err)
	}
	return p.(int)
}

func (c *Container) ParamB() int {
	p, err := c.GetParam("b")
	if err != nil {
		panic(err)
	}
	return p.(int)
}

func describeEndpointCompareParams() service.Service {
	e := service.New()
	e.SetConstructor(newEndpointCompareParams, dependency.Service("container"))
	return e
}

func describeMux() service.Service {
	m := container.NewService()
	m.
		SetConstructor(containerHttp.NewServeMux, dependency.Container()).
		AppendCall("HandleDynamic", dependency.Value("/"), dependency.Value("endpointCompareParams"))
	return m
}

func NewContainer() *Container {
	c := &Container{container.New()}

	root := service.New()
	root.SetConstructor(func() any {
		return c
	})

	c.OverrideServices(service.Services{
		"httpHandler":           describeMux(),
		"endpointCompareParams": describeEndpointCompareParams(),
		"container":             root,
	})
	c.OverrideParams(param.Params{
		"a": dependency.Value(0),
		"b": dependency.Value(0),
	})

	return c
}

// params is an interface that is implemented by [*Container]
// we inject it to [newEndpointCompareParams]
// "Clients should not be forced to depend upon interfaces that they do not use"
type params interface {
	ParamA() int
	ParamB() int
}

// newEndpointCompareParams helps us to check whether changing two parameters
// using HotSwap is an atomic operation.
func newEndpointCompareParams(p params) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p.ParamA() == p.ParamB() {
			_, _ = w.Write([]byte("p.ParamA() == p.ParamB()"))
		} else {
			_, _ = w.Write([]byte("p.ParamA() != p.ParamB()"))
		}
	})
}
