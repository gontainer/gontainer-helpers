# Container

This package provides a concurrently safe DI container. It supports **scoped variables**, and **atomic hot swapping**.
For bigger projects, it provides a tool for the code generation.

See [docs](docs).

```bash
go get -u github.com/gontainer/gontainer-helpers/v3/container@latest
```

## Examples

### HotSwap

```go
package examples

import (
	"database/sql"
	"time"

	"github.com/gontainer/gontainer-helpers/v3/container"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/service"
)

func RefreshConfigEveryMinute(c *container.Container) {
	go func () {
		for {
			<-time.After(time.Minute)

			// HotSwap guarantees atomicity
			c.HotSwap(func(c container.MutableContainer) {
				// override the value of a param
				// the cache for that param is automatically invalidated
				c.OverrideParam("my-param", container.NewDependencyValue(125))

				// override a service
				// the cache for that service is automatically invalidated
				db := service.New()
				db.SetConstructor(
					sql.Open,
					container.NewDependencyValue("mysql"),
					container.NewDependencyParam("dataSourceName"),
				)

				// invalidate the cache for the given params...
				c.InvalidateParamsCache("paramA", "paramB")
				// ... or for all of them
				c.InvalidateAllParamsCache()

				// invalidate the cache for the given service...
				c.InvalidateServicesCache("serviceA", "serviceB")
				/// or for all of them
				c.InvalidateAllServicesCache()
			})
		}
	}()
}
```

### Dependency injection

```go
package examples

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/v3/container"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/dependency"
	"github.com/gontainer/gontainer-helpers/v3/container/shortcuts/service"
)

type God struct {
	Name string
}

func NewGod(name string) God {
	return God{Name: name}
}

type MountOlympus struct {
	Gods []God
}

func buildContainer() *container.Container {
	// describe mountOlympus
	mountOlympus := service.New()
	mountOlympus.
		SetValue(MountOlympus{}).
		SetField("Gods", dependency.Tag("olympians"))

	// describe athena
	athena := service.New()
	athena.
		SetValue(God{Name: "Athena"}).
		Tag("olympians", 1) // priority 1 - ladies first :)

	// describe zeus
	zeus := service.New()
	zeus.
		SetConstructor(NewGod, dependency.Value("Zeus")). // constructor injection
		Tag("olympians", 0)
	
	c := container.New()
	c.OverrideServices(service.Services{
		"mountOlympus": mountOlympus,
		"athena":       athena,
		"zeus":         zeus,
	})

	return c
}

func Example() {
	c := buildContainer()
	mountOlympus, _ := c.Get("mountOlympus")
	fmt.Println(mountOlympus)
	// Output: {[{Athena} {Zeus}]}
}
```