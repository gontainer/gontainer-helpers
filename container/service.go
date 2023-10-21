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

type Service struct {
	value           any
	constructor     any
	constructorDeps []Dependency
	calls           []serviceCall
	fields          []serviceField
	tags            map[string]int
	scope           scope
}

func NewService() Service {
	return Service{
		tags:  make(map[string]int),
		scope: scopeDefault,
	}
}

func (s *Service) SetValue(v any) *Service {
	s.value = v
	return s
}

func (s *Service) SetConstructor(fn any, deps ...Dependency) *Service {
	s.constructor = fn
	s.constructorDeps = deps
	return s
}

func (s *Service) AppendCall(method string, deps ...Dependency) *Service {
	s.calls = append(s.calls, serviceCall{
		wither: false,
		method: method,
		deps:   deps,
	})
	return s
}

func (s *Service) AppendWither(method string, deps ...Dependency) *Service {
	s.calls = append(s.calls, serviceCall{
		wither: true,
		method: method,
		deps:   deps,
	})
	return s
}

func (s *Service) SetField(field string, dep Dependency) *Service {
	s.fields = append(s.fields, serviceField{
		name: field,
		dep:  dep,
	})
	return s
}

func (s *Service) Tag(tag string, priority int) *Service {
	s.tags[tag] = priority
	return s
}

func (s *Service) SetScopeDefault() *Service {
	s.scope = scopeDefault
	return s
}

func (s *Service) SetScopeShared() *Service {
	s.scope = scopeShared
	return s
}

func (s *Service) SetScopeContextual() *Service {
	s.scope = scopeContextual
	return s
}

func (s *Service) SetScopeNonShared() *Service {
	s.scope = scopeNonShared
	return s
}
