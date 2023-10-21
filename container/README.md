# Container

This package provides a concurrently safe DI container. It supports scoped variables, and hot swapping.

## Quick start

```go
package main

import (
	"database/sql"
	"fmt"

	"github.com/gontainer/gontainer-helpers/container"
)

func buildContainer() *container.Container {
	// describe db
	db := container.NewService()
	db.SetConstructor(
		func(username, password string) (*sql.DB, error) {
			return sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/test", username, password))
		},
		container.NewDependencyParam("db.user"),
		container.NewDependencyParam("db.password"),
	)

	// register dependencies
	c := container.New()
	c.OverrideService("db", db)
	c.OverrideParam("db.user", container.NewDependencyValue("root"))
	c.OverrideParam("db.password", container.NewDependencyValue("root"))

	return c
}

func main() {
	c := buildContainer()
	db, err := c.Get("db")
	// ... more code
}
```
