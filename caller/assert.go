package caller

import (
	"github.com/gontainer/gontainer-helpers/internal/caller"
)

func AssertFunc(fn interface{}) {
	if _, err := caller.Func(fn); err != nil {
		panic(err)
	}
}

func AssertProvider(fn interface{}) {
	if _, err := caller.FuncProvider(fn); err != nil {
		panic(err)
	}
}

func AssertMethod(object interface{}, method string) {
	if _, err := caller.Method(object, method); err != nil {
		panic(err)
	}
}

func AssertWither(object interface{}, method string) {
	if _, err := caller.Wither(object, method); err != nil {
		panic(err)
	}
}
