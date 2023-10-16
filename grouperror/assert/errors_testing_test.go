package assert_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/gontainer/gontainer-helpers/grouperror"
	errAssert "github.com/gontainer/gontainer-helpers/grouperror/assert"
	"github.com/stretchr/testify/assert"
)

type mockTesting string

func (m *mockTesting) Errorf(format string, args ...interface{}) {
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
