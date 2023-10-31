// Copyright (c) 2023 Bartłomiej Krukowski
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is furnished
// to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

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
