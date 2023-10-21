package container

type scope uint

const (
	scopeDefault    scope = iota // scopeShared if has a contextual dependency, otherwise scopeShared
	scopeShared                  // The same instance is used each time you request it from the Container
	scopeContextual              // The same instance is shared only between the created Service and its direct and indirect dependencies
	scopeNonShared               // New instance is created each time you request it from the Container
)

var scopeNames = map[scope]string{
	scopeDefault:    "scopeDefault",
	scopeShared:     "scopeShared",
	scopeContextual: "scopeContextual",
	scopeNonShared:  "scopeNonShared",
}

func (s scope) String() string {
	if str, ok := scopeNames[s]; ok {
		return str
	}
	return "unknown"
}
