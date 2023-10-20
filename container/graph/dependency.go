package graph

type dependencyKind uint

const (
	dependencyDecoratedByTag dependencyKind = iota + 1
	dependencyTag
	dependencyService
	dependencyParam
	dependencyDecorator
)

type Dependency struct {
	id   string // unique identifier in the graph
	kind dependencyKind

	Resource string // name of service, name of tag or decoratorID
	Pretty   string // pretty name
}

func (d Dependency) IsService() bool {
	return d.kind == dependencyService
}
