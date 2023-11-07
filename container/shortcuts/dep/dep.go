package dep

import (
	"github.com/gontainer/gontainer-helpers/v3/container"
)

type (
	Dependency = container.Dependency
)

var (
	Value     = container.NewDependencyValue
	Tag       = container.NewDependencyTag
	Service   = container.NewDependencyService
	Param     = container.NewDependencyParam
	Provider  = container.NewDependencyProvider
	Container = container.NewDependencyContainer
	Context   = container.NewDependencyContext
)
