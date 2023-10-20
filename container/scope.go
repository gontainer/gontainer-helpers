package container

type scope uint

const (
	scopeDefault    scope = iota // scopeShared if has a contextual dependency, otherwise scopeShared
	scopeShared                  // The same instance is used each time you request it from the container
	scopeContextual              // The same instance is shared only between the created Service and its direct and indirect dependencies
	scopeNonShared               // New instance is created each time you request it from the container
)

var scopeStringMapping = map[scope]string{
	scopeDefault:    "scopeDefault",
	scopeShared:     "scopeShared",
	scopeContextual: "scopeContextual",
	scopeNonShared:  "scopeNonShared",
}

func (s scope) String() string {
	if str, ok := scopeStringMapping[s]; ok {
		return str
	}
	return "unknown"
}
