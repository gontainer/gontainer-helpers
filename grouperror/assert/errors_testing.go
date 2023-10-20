package assert

import (
	"github.com/gontainer/gontainer-helpers/grouperror"
	"github.com/stretchr/testify/assert"
)

// testingT is an interface wrapper around *testing.T
type testingT interface {
	Errorf(format string, args ...any)
}

func minLen(a []error, b []string) int {
	x := len(a)
	y := len(b)

	if x < y {
		return x
	}

	return y
}

// EqualErrorGroup asserts that the given error is a group of errors with the following messages.
// It asserts the given error is equal to nil whenever `len(msgs) == 0`.
//
// see grouperror.Collection
func EqualErrorGroup(t testingT, err error, msgs []string) {
	if len(msgs) == 0 {
		assert.NoError(t, err)
		return
	}

	errs := grouperror.Collection(err)
	l := minLen(errs, msgs)
	for i := 0; i < l; i++ {
		assert.EqualError(t, errs[i], msgs[i])
	}

	var extra []any
	if err != nil {
		extra = []any{err.Error()}
	}

	assert.Len(t, errs, len(msgs), extra...)
}
