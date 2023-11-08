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
	"github.com/gontainer/gontainer-helpers/v3/container/internal/examples/testserver/repositories"
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

func describeUserRepository() service.Service {
	// ur := repositories.ImageRepository{}
	// ur.Tx = c.Get("tx")
	s := service.New()
	s.
		SetValue(repositories.UserRepository{}).
		SetField("Tx", dependency.Service("tx"))
	return s
}

func describeImageRepository() service.Service {
	// ir := repositories.ImageRepository{}
	// ir.Tx = c.Get("tx")
	s := service.New()
	s.
		SetValue(repositories.ImageRepository{}).
		SetField("Tx", dependency.Service("tx"))
	return s
}

func describeMyEndpoint() service.Service {
	s := service.New()
	s.
		SetConstructor(pkgHttp.NewMyEndpoint, dependency.Service("rootContainer")).
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
	c.OverrideServices(map[string]service.Service{
		"db": describeDB(),
		"tx": describeTx(),

		"userRepo":  describeUserRepository(),
		"imageRepo": describeImageRepository(),

		"mux":        describeMux(),
		"server":     describeServer(),
		"myEndpoint": describeMyEndpoint(),
	})

	/*
		Container embeds *container.Container and adds custom methods, e.g.:
		`Tx(ctx context.Context) *sql.Tx`
		Registering it as a service lets us inject it in our handlers, e.g.:

		type TxFactory interface {
			Tx(context.Context) *sql.Tx
		}

		func NewMyHandler(f TxFactory) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tx := f.Tx(r.Context()) // fetch the transaction for the given context
				// ... your code
			})
		}
	*/
	rootContainer := service.New()
	rootContainer.SetConstructor(func() interface{} {
		return c
	})
	c.OverrideService("rootContainer", rootContainer)

	/*
		The following decorator it's an equivalent of the following code:
		var tmp pkgHttp.ErrorAwareHandler
		tmp := ... // your code
		h := pkgHttp.NewAutoCloseTxEndpoint(c, tmp)
	*/
	c.AddDecorator(
		"error-aware-handler",
		func(p container.DecoratorPayload, c *Container) http.Handler {
			return pkgHttp.NewAutoCloseTxEndpoint(c, p.Service.(pkgHttp.ErrorAwareHandler))
		},
		dependency.Service("rootContainer"),
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

type Container struct {
	*container.Container
}

func (c *Container) mustGet(ctx context.Context, n string) interface{} {
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

func (c *Container) DB(ctx context.Context) *sql.DB {
	return c.mustGet(ctx, "db").(*sql.DB)
}

func (c *Container) Tx(ctx context.Context) *sql.Tx {
	return c.mustGet(ctx, "tx").(*sql.Tx)
}

func (c *Container) UserRepo(ctx context.Context) repositories.UserRepository {
	return c.mustGet(ctx, "userRepo").(repositories.UserRepository)
}

func (c *Container) ImageRepo(ctx context.Context) repositories.ImageRepository {
	return c.mustGet(ctx, "imageRepo").(repositories.ImageRepository)
}

func (c *Container) Mux(ctx context.Context) *containerHttp.ServeMux {
	return c.mustGet(ctx, "mux").(*containerHttp.ServeMux)
}

func (c *Container) Server(ctx context.Context) *http.Server {
	return c.mustGet(ctx, "server").(*http.Server)
}

func (c *Container) ServerAddr() string {
	return c.mustGetParam("SERVER_ADDR").(string)
}
