package container

// TODO: remove it
type SuperContainer struct {
	*container
	*paramContainer
}

// NewSuperContainer creates a concurrent-safe DI container.
//
// TODO: remove it
func NewSuperContainer() *SuperContainer {
	return &SuperContainer{
		container:      NewContainer(),
		paramContainer: NewParamContainer(),
	}
}
