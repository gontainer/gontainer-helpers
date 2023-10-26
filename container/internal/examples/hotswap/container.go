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
