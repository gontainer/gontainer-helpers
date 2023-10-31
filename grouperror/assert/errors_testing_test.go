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

package assert_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/gontainer/gontainer-helpers/v2/grouperror"
	errAssert "github.com/gontainer/gontainer-helpers/v2/grouperror/assert"
	"github.com/stretchr/testify/assert"
)

type mockTesting string

func (m *mockTesting) Errorf(format string, args ...any) {
	*m += mockTesting(fmt.Sprintf(format, args...))
}

func (m *mockTesting) String() string {
	return string(*m)
}

func TestEqualErrorGroup(t *testing.T) {
	t.Run("No errors [OK]", func(t *testing.T) {
		mt := new(mockTesting)
		errAssert.EqualErrorGroup(mt, nil, nil)
		assert.Empty(t, mt.String())
	})

	t.Run("No errors [error]", func(t *testing.T) {
		mt := new(mockTesting)
		errAssert.EqualErrorGroup(mt, os.ErrClosed, nil)
		assert.Equal(
			t,
			`
	Error Trace:	
	Error:      	Received unexpected error:
	            	file already closed
`,
			mt.String(),
		)
	})

	t.Run("Equal errors [OK]", func(t *testing.T) {
		mt := new(mockTesting)
		errAssert.EqualErrorGroup(
			mt,
			grouperror.Join(os.ErrClosed, os.ErrExist),
			[]string{
				"file already closed",
				"file already exists",
			},
		)
		assert.Empty(t, mt.String())
	})

	t.Run("Equal errors [error #1]", func(t *testing.T) {
		mt := new(mockTesting)
		errAssert.EqualErrorGroup(
			mt,
			grouperror.Join(os.ErrClosed, os.ErrExist),
			[]string{
				"file already closed",
			},
		)
		assert.Equal(
			t,
			`
	Error Trace:	
	Error:      	"[file already closed file already exists]" should have 1 item(s), but has 2
	Messages:   	file already closed
	            	file already exists
`,
			mt.String(),
		)
	})

	t.Run("Equal errors [error #2]", func(t *testing.T) {
		mt := new(mockTesting)
		errAssert.EqualErrorGroup(
			mt,
			grouperror.Join(os.ErrClosed),
			[]string{
				"file already closed",
				"file already exists",
			},
		)
		assert.Equal(
			t,
			`
	Error Trace:	
	Error:      	"[file already closed]" should have 2 item(s), but has 1
	Messages:   	file already closed
`,
			mt.String(),
		)
	})

	t.Run("Equal errors [error #3]", func(t *testing.T) {
		mt := new(mockTesting)
		errAssert.EqualErrorGroup(
			mt,
			grouperror.Join(os.ErrClosed, os.ErrExist),
			[]string{
				"file already exists",
				"file already closed",
			},
		)
		assert.Equal(
			t,
			`
	Error Trace:	
	Error:      	Error message not equal:
	            	expected: "file already exists"
	            	actual  : "file already closed"

	Error Trace:	
	Error:      	Error message not equal:
	            	expected: "file already closed"
	            	actual  : "file already exists"
`,
			mt.String(),
		)
	})

	t.Run("Equal errors [error #4]", func(t *testing.T) {
		mt := new(mockTesting)
		errAssert.EqualErrorGroup(
			mt,
			grouperror.Join(os.ErrClosed, os.ErrInvalid, os.ErrExist),
			[]string{
				"file already exists",
				"file already closed",
			},
		)
		assert.Equal(
			t,
			`
	Error Trace:	
	Error:      	Error message not equal:
	            	expected: "file already exists"
	            	actual  : "file already closed"

	Error Trace:	
	Error:      	Error message not equal:
	            	expected: "file already closed"
	            	actual  : "invalid argument"

	Error Trace:	
	Error:      	"[file already closed invalid argument file already exists]" should have 2 item(s), but has 3
	Messages:   	file already closed
	            	invalid argument
	            	file already exists
`,
			mt.String(),
		)
	})

	t.Run("Equal errors [error #5]", func(t *testing.T) {
		mt := new(mockTesting)
		errAssert.EqualErrorGroup(
			mt,
			grouperror.Join(os.ErrClosed, os.ErrInvalid),
			[]string{
				"invalid argument",
				"file already exists",
				"file already closed",
			},
		)
		assert.Equal(
			t,
			`
	Error Trace:	
	Error:      	Error message not equal:
	            	expected: "invalid argument"
	            	actual  : "file already closed"

	Error Trace:	
	Error:      	Error message not equal:
	            	expected: "file already exists"
	            	actual  : "invalid argument"

	Error Trace:	
	Error:      	"[file already closed invalid argument]" should have 3 item(s), but has 2
	Messages:   	file already closed
	            	invalid argument
`,
			mt.String(),
		)
	})
}
