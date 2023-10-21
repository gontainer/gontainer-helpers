package hotswap

import (
	"fmt"
	"net/http"

	"github.com/gontainer/gontainer-helpers/v2/container"
)

type Container struct {
	*container.Container
}

func (c *Container) ServeMux() *http.ServeMux {
	s, err := c.Get("mux")
	if err != nil {
		panic(err)
	}
	return s.(*http.ServeMux)
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

	c.OverrideService("mux", m)
	c.OverrideParam("a", container.NewDependencyValue(0))
	c.OverrideParam("b", container.NewDependencyValue(0))

	return c
}

func newHandleHomePage(c *Container) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// assign context to the container
		ctx := container.ContextWithContainer(r.Context(), c)
		// uncomment the following line if you want to pass the request to another func
		// r := r.Clone(ctx)

		_ = ctx

		_, _ = w.Write([]byte(fmt.Sprintf("%d=%d", c.ParamA(), c.ParamB())))
	})
}
