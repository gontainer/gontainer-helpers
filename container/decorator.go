package container

type DecoratorContext struct {
	Tag       string
	ServiceID string
	Service   interface{}
}
