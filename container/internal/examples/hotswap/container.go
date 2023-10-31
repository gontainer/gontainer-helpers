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

package hotswap

import (
	"net/http"

	"github.com/gontainer/gontainer-helpers/v2/container"
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

func NewContainer() *Container {
	c := &Container{container.New()}

	m := container.NewService()
	m.SetConstructor(http.NewServeMux)
	m.AppendCall(
		"Handle",
		container.NewDependencyValue("/"),
		container.NewDependencyValue(newHandleHomePage(c)),
	)
	m.Tag("http-handler", 0)

	c.OverrideService("httpHandler", m)
	c.OverrideParam("a", container.NewDependencyValue(0))
	c.OverrideParam("b", container.NewDependencyValue(0))

	c.AddDecorator(
		"http-handler",
		decorateHandlerByContainer,
		container.NewDependencyContainer(),
	)

	return c
}

func decorateHandlerByContainer(p container.DecoratorPayload, c *container.Container) http.Handler {
	return container.HTTPHandlerWithContainer(p.Service.(http.Handler), c)
}

func newHandleHomePage(c *Container) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c.ParamA() == c.ParamB() {
			_, _ = w.Write([]byte("c.ParamA() == c.ParamB()"))
		} else {
			_, _ = w.Write([]byte("c.ParamA() != c.ParamB()"))
		}
	})
}
