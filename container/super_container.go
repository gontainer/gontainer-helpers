package container

type SuperContainer struct {
	*container
	*paramContainer
}

// NewSuperContainer creates a concurrent-safe DI container.
func NewSuperContainer() *SuperContainer {
	return &SuperContainer{
		container:      NewContainer(),
		paramContainer: NewParamContainer(),
	}
}
