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

	h := container.NewService()
	h.SetConstructor(
		container.HTTPHandlerWithContainer,
		container.NewDependencyService("serveMux"),
		container.NewDependencyValue(c),
	)

	c.OverrideService("serveMux", m)
	c.OverrideService("httpHandler", h)
	c.OverrideParam("a", container.NewDependencyValue(0))
	c.OverrideParam("b", container.NewDependencyValue(0))

	return c
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
