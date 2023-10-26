package container

type dependencyType int

const (
	dependencyMissing dependencyType = iota // zero-value, invalid value
	dependencyValue
	dependencyTag
	dependencyService
	dependencyParam
	dependencyProvider
	dependencyContainer
)

var dependencyNames = map[dependencyType]string{
	dependencyMissing:   "dependencyMissing",
	dependencyValue:     "dependencyValue",
	dependencyTag:       "dependencyTag",
	dependencyService:   "dependencyService",
	dependencyParam:     "dependencyParam",
	dependencyProvider:  "dependencyProvider",
	dependencyContainer: "dependencyContainer",
}

func (d dependencyType) String() string {
	if s, ok := dependencyNames[d]; ok {
		return s
	}
	return "unknown"
}

/*
Dependency represents a dependency in a [*Container].
Use on of the following func to create a new one:
  - [NewDependencyValue]
  - [NewDependencyTag]
  - [NewDependencyService]
  - [NewDependencyParam]
  - [NewDependencyProvider]
  - [NewDependencyContainer]
*/
type Dependency struct {
	type_     dependencyType
	value     any
	tagID     string
	serviceID string
	paramID   string
	provider  any
}

// NewDependencyValue creates a value-[Dependency], it does not depend on anything in a [*Container].
func NewDependencyValue(v any) Dependency {
	return Dependency{
		type_: dependencyValue,
		value: v,
	}
}

// NewDependencyTag creates a [Dependency] to the given tag
func NewDependencyTag(tagID string) Dependency {
	return Dependency{
		type_: dependencyTag,
		tagID: tagID,
	}
}

// NewDependencyService creates a [Dependency] to the given Service
func NewDependencyService(serviceID string) Dependency {
	return Dependency{
		type_:     dependencyService,
		serviceID: serviceID,
	}
}

// NewDependencyParam creates a [Dependency] to the given parameter.
func NewDependencyParam(paramID string) Dependency {
	return Dependency{
		type_:   dependencyParam,
		paramID: paramID,
	}
}

// NewDependencyProvider creates a [Dependency] that will be returned by the given provider.
func NewDependencyProvider(provider any) Dependency {
	return Dependency{
		type_:    dependencyProvider,
		provider: provider,
	}
}

// NewDependencyContainer creates a [Dependency] to the [*Container].
func NewDependencyContainer() Dependency {
	return Dependency{
		type_: dependencyContainer,
	}
}
