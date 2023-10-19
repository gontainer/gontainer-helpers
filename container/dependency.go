package container

type dependencyType int

const (
	dependencyMissing dependencyType = iota // zero-value, invalid value
	dependencyNil
	dependencyTag
	dependencyService
	dependencyProvider
)

var dependencyStringMapping = map[dependencyType]string{
	dependencyMissing:  "dependencyMissing",
	dependencyNil:      "dependencyNil",
	dependencyTag:      "dependencyTag",
	dependencyService:  "dependencyService",
	dependencyProvider: "dependencyProvider",
}

func (d dependencyType) String() string {
	if s, ok := dependencyStringMapping[d]; ok {
		return s
	}
	return "unknown"
}

type Dependency struct {
	type_     dependencyType
	value     any
	tagID     string
	serviceID string
	provider  any
}

// NewDependencyValue creates a nil-Dependency, it does not depend on anything in the container
func NewDependencyValue(v any) Dependency {
	return Dependency{
		type_: dependencyNil,
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

func NewDependencyProvider(provider any) Dependency {
	return Dependency{
		type_:    dependencyProvider,
		provider: provider,
	}
}
