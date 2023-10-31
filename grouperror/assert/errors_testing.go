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

package assert

import (
	"github.com/gontainer/gontainer-helpers/v2/grouperror"
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
