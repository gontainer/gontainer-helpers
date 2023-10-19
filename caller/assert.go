package caller

import (
	"github.com/gontainer/gontainer-helpers/internal/caller"
)

// TODO: do we need it? do we need a second argument to check the number of args?

// AssertFunc panics if the given input is not a func.
func AssertFunc(fn any) {
	if _, err := caller.Func(fn); err != nil {
		panic(err)
	}
}

// AssertProvider panics if the given input is not a provider.
func AssertProvider(fn any) {
	if _, err := caller.FuncProvider(fn); err != nil {
		panic(err)
	}
}

// AssertMethod panics if the given input is not a method.
func AssertMethod(object any, method string) {
	if _, err := caller.Method(object, method); err != nil {
		panic(err)
	}
}

// AssertWither panics if the given input is not a wither.
func AssertWither(object any, method string) {
	if _, err := caller.Wither(object, method); err != nil {
		panic(err)
	}
}
