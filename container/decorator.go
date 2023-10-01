package container

type DecoratorPayload struct {
	Tag       string
	ServiceID string
	Service   interface{}
}

// DecoratorContext an alias to DecoratorPayload.
//
// Deprecated: that name may sound confusing.
type DecoratorContext = DecoratorPayload
