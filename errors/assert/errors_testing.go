package assert

import (
	"github.com/gontainer/gontainer-helpers/errors"
	"github.com/stretchr/testify/assert"
)

// testingT is an interface wrapper around *testing.T
type testingT interface {
	Errorf(format string, args ...interface{})
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
// see errors.Collection
func EqualErrorGroup(t testingT, err error, msgs []string) {
	if len(msgs) == 0 {
		assert.NoError(t, err)
		return
	}

	errs := errors.Collection(err)
	for i := 0; i < minLen(errs, msgs); i++ {
		assert.EqualError(t, errs[i], msgs[i])
	}

	var extra []interface{}
	if err != nil {
		extra = append(extra, err.Error())
	}

	assert.Len(t, errs, len(msgs), extra...)
}
