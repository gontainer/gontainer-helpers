package container

type dependencyType int

const (
	dependencyMissing dependencyType = iota // zero-value, invalid value
	dependencyValue
	dependencyTag
	dependencyService
	dependencyParam
	dependencyProvider
)

var dependencyNames = map[dependencyType]string{
	dependencyMissing:  "dependencyMissing",
	dependencyValue:    "dependencyValue",
	dependencyTag:      "dependencyTag",
	dependencyService:  "dependencyService",
	dependencyParam:    "dependencyParam",
	dependencyProvider: "dependencyProvider",
}

func (d dependencyType) String() string {
	if s, ok := dependencyNames[d]; ok {
		return s
	}
	return "unknown"
}

type Dependency struct {
	type_     dependencyType
	value     any
	tagID     string
	serviceID string
	paramID   string
	provider  any
}

// NewDependencyValue creates a nil-Dependency, it does not depend on anything in the Container
func NewDependencyValue(v any) Dependency {
	return Dependency{
		type_: dependencyValue,
		value: v,
	}
}

// NewDependencyTag creates a Dependency to the given tag
func NewDependencyTag(tagID string) Dependency {
	return Dependency{
		type_: dependencyTag,
		tagID: tagID,
	}
}

// NewDependencyService creates a Dependency to the given Service
func NewDependencyService(serviceID string) Dependency {
	return Dependency{
		type_:     dependencyService,
		serviceID: serviceID,
	}
}

func NewDependencyParam(paramID string) Dependency {
	return Dependency{
		type_:   dependencyParam,
		paramID: paramID,
	}
}

func NewDependencyProvider(provider any) Dependency {
	return Dependency{
		type_:    dependencyProvider,
		provider: provider,
	}
}
