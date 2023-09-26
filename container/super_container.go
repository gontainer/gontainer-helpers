package container

type SuperContainer struct {
	*container
	*paramContainer
}

func NewSuperContainer() *SuperContainer {
	return &SuperContainer{
		container:      NewContainer(),
		paramContainer: NewParamContainer(),
	}
}
