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
	"context"
	"database/sql"
	"net/http"
	"os"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gontainer/gontainer-helpers/v3/container"
	containerHttp "github.com/gontainer/gontainer-helpers/v3/container/http"
	pkgHttp "github.com/gontainer/gontainer-helpers/v3/container/internal/examples/testserver/http"
	"github.com/gontainer/gontainer-helpers/v3/container/internal/examples/testserver/repos"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/dependency"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/field"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/service"
)

func describeDB() service.Service {
	s := service.New()
	s.SetConstructor(func() (*sql.DB, error) {
		// use the mock db to avoid a dependency on a real db
		db, mock, err := sqlmock.New()

		// required for test to avoid the following error:
		// all expectations were already fulfilled, call to database transaction Begin was not expected
		for i := 0; i < 5000; i++ {
			mock.ExpectBegin()
			mock.ExpectCommit()
		}

		return db, err
	})
	return s
}

func describeTx() service.Service {
	s := service.New()
	// tx, err := db.BeginTx(ctx, nil)
	s.
		SetFactory("db", "BeginTx", dependency.Context(), dependency.Value(nil)).
		SetScopeContextual() // IMPORTANT
	// SetScopeContextual instructs the container to create a new instance of that service for each context
	return s
}

func describeUserRepo() service.Service {
	// ur := repos.ImageRepo{}
	// ur.Tx = c.Get("tx")
	s := service.New()
	s.
		SetValue(repos.UserRepo{}).
		SetField("Tx", dependency.Service("tx"))
	// NOTE
	// userRepo has the contextual scope automatically,
	// because it depends on the "tx" service that has the contextual scope
	return s
}

func describeImageRepo() service.Service {
	// ir := repos.ImageRepo{}
	// ir.Tx = c.Get("tx")
	s := service.New()
	s.
		SetValue(repos.ImageRepo{}).
		SetField("Tx", dependency.Service("tx"))
	// NOTE
	// imageRepo has the contextual scope automatically,
	// because it depends on the "tx" service that has the contextual scope
	return s
}

func describeMyEndpoint() service.Service {
	// myEndpoint := pkgHttp.NewMyEndpoint(c.Get("userRepo"), c.Get("imageRepo"))
	s := service.New()
	s.
		SetConstructor(pkgHttp.NewMyEndpoint, dependency.Service("userRepo"), dependency.Service("imageRepo")).
		Tag("error-aware-handler", 0)
	return s
}

func describeServer() service.Service {
	s := service.New()
	s.SetConstructor(func() *http.Server {
		return &http.Server{}
	})
	s.SetFields(field.Fields{
		"Addr":    dependency.Param("SERVER_ADDR"), // fetch the server address from the param to make it configurable
		"Handler": dependency.Service("mux"),
	})
	return s
}

func describeMux() service.Service {
	s := service.New()
	s.
		SetConstructor(containerHttp.NewServeMux, dependency.Container()).
		AppendCall("HandleDynamic", dependency.Value("/"), dependency.Value("myEndpoint"))
	return s
}

func BuildContainer() *Container {
	c := &Container{container.New()}
	c.OverrideServices(service.Services{
		"db": describeDB(),
		"tx": describeTx(),

		"userRepo":  describeUserRepo(),
		"imageRepo": describeImageRepo(),

		"mux":        describeMux(),
		"server":     describeServer(),
		"myEndpoint": describeMyEndpoint(),
	})

	/*
		The following decorator is an equivalent of the following code:

		var tmp pkgHttp.ErrorAwareHandler
		tmp := ... // your code
		h := pkgHttp.NewAutoCloseTxEndpoint(tx, tmp)
	*/
	c.AddDecorator(
		"error-aware-handler",
		func(p container.DecoratorPayload, tx *sql.Tx) http.Handler {
			return pkgHttp.NewAutoCloseTxEndpoint(tx, p.Service.(pkgHttp.ErrorAwareHandler))
		},
		dependency.Service("tx"),
	)

	// make the server address configurable
	c.OverrideParam("SERVER_ADDR", dependency.Provider(func() string {
		addr := os.Getenv("SERVER_ADDR")
		if addr == "" {
			addr = ":8080"
		}
		return addr
	}))

	return c
}

// Container wraps [*container.Container].
// This approach lets us create custom getters, see:
//   - [*Container.Server]
//   - [*Container.ServerAddr]
type Container struct {
	*container.Container
}

func (c *Container) mustGet(ctx context.Context, n string) any {
	v, err := c.GetInContext(ctx, n)
	if err != nil {
		panic(err)
	}
	return v
}

func (c *Container) mustGetParam(n string) interface{} {
	v, err := c.GetParam(n)
	if err != nil {
		panic(err)
	}
	return v
}

func (c *Container) Server(ctx context.Context) *http.Server {
	return c.mustGet(ctx, "server").(*http.Server)
}

func (c *Container) ServerAddr() string {
	return c.mustGetParam("SERVER_ADDR").(string)
}
