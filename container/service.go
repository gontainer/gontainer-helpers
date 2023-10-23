package container

type serviceCall struct {
	wither bool
	method string
	deps   []Dependency
}

type serviceField struct {
	name string
	dep  Dependency
}

// Service represents a service in the [*Container].
// Use [NewService] to create a new instance.
type Service struct {
	hasCreationMethod bool
	value             any
	constructor       any
	constructorDeps   []Dependency
	calls             []serviceCall
	fields            []serviceField
	tags              map[string]int
	scope             scope
}

// NewService creates a new service.
func NewService() Service {
	return Service{
		tags:  make(map[string]int),
		scope: scopeDefault,
	}
}

func (s *Service) resetCreationMethods() {
	s.value = nil
	s.constructor = nil
	s.constructorDeps = nil
}

// SetValue sets a predefined value of the service. It excludes [*Service.SetConstructor].
func (s *Service) SetValue(v any) *Service {
	s.resetCreationMethods()
	s.value = v
	s.hasCreationMethod = true
	return s
}

/*
SetConstructor sets a constructor of the service. It excludes [*Service.SetValue].

	s := container.NewService()
	s.SetConstructor(
		func NewDB(username, password string) (*sql.DB, error) {
			return sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/test", username, password))
		},
		container.NewDependencyValue("root"),
		container.NewDependencyValue("root"),
	)
*/
func (s *Service) SetConstructor(fn any, deps ...Dependency) *Service {
	s.resetCreationMethods()
	s.constructor = fn
	s.constructorDeps = deps
	s.hasCreationMethod = true
	return s
}

/*
AppendCall instructs the container to execute a method over that object.

	s := container.NewService()
	s.SetConstructor(func() *Person {
		return &Person{}
	})
	s.AppendMethod("SetName", container.NewDependencyValue("Jane"))

	// p := &Person{}
	// p.SetName("Jane")
*/
func (s *Service) AppendCall(method string, deps ...Dependency) *Service {
	s.calls = append(s.calls, serviceCall{
		wither: false,
		method: method,
		deps:   deps,
	})
	return s
}

/*
AppendWither instructs the container to execute a wither-method over that object.

	s := container.NewService()
	s.SetValue(Person{})
	s.AppendWither("WithName", container.NewDependencyValue("Jane"))

	// p := Person{}
	// p = p.WithName("Jane")
*/
func (s *Service) AppendWither(method string, deps ...Dependency) *Service {
	s.calls = append(s.calls, serviceCall{
		wither: true,
		method: method,
		deps:   deps,
	})
	return s
}

/*
SetField instructs the container to set a value of the given field.

	s := container.NewService()
	s.SetValue(Person{})
	s.SetField("Name", container.NewDependencyValue("Jane"))

	// p := Person{}
	// p.Name = "Jane"
*/
func (s *Service) SetField(field string, dep Dependency) *Service {
	s.fields = append(s.fields, serviceField{
		name: field,
		dep:  dep,
	})
	return s
}

// Tag tags the given service. Argument priority it is being used for determining order in [*Container.GetTaggedBy].
func (s *Service) Tag(tag string, priority int) *Service {
	s.tags[tag] = priority
	return s
}

// SetScopeDefault sets the scope that will be determined in real-time.
func (s *Service) SetScopeDefault() *Service {
	s.scope = scopeDefault
	return s
}

// SetScopeShared sets the shared scope.
func (s *Service) SetScopeShared() *Service {
	s.scope = scopeShared
	return s
}

// SetScopeContextual sets the contextual scope.
//
// See [*Container.GetInContext].
func (s *Service) SetScopeContextual() *Service {
	s.scope = scopeContextual
	return s
}

// SetScopeNonShared sets the non-shared scope.
// Each invocation of [*Container.Get] will create a new variable.
func (s *Service) SetScopeNonShared() *Service {
	s.scope = scopeNonShared
	return s
}
