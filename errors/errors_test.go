package errors_test

import (
	pkgErrors "errors"
	"fmt"
	"io"
	"runtime"
	"strings"
	"testing"

	"github.com/gontainer/gontainer-helpers/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrefixedGroup(t *testing.T) {
	t.Parallel()

	t.Run("Empty", func(t *testing.T) {
		assert.NoError(t, errors.Group(nil, nil))
	})

	t.Run("One-error collection", func(t *testing.T) {
		errs := errors.Collection(io.EOF)
		assert.Equal(t, []error{io.EOF}, errs)
	})

	t.Run("Errors", func(t *testing.T) {
		fieldsErr := errors.Group(
			fmt.Errorf("invalid value of `name`"),
			fmt.Errorf("invalid value of `age`"),
		)

		personErr := errors.PrefixedGroup(
			"Person: ",
			fieldsErr,
			fmt.Errorf("the given ID does not exist"),
		)
		validationErr := errors.PrefixedGroup(
			"Validation: ",
			personErr,
			fmt.Errorf("unexpected error"),
		)

		errs := errors.Collection(validationErr)
		expected := []string{
			"Validation: Person: invalid value of `name`",
			"Validation: Person: invalid value of `age`",
			"Validation: Person: the given ID does not exist",
			"Validation: unexpected error",
		}

		require.Len(t, errs, len(expected), validationErr)
		for i, err := range errs {
			assert.EqualError(t, err, expected[i])
		}

		assert.EqualError(t, validationErr, strings.Join(expected, "\n"))
	})

	t.Run("Unwrap()", func(t *testing.T) {
		// Go 1.20 expands support for error wrapping to permit an error to wrap multiple other errors.
		// An error e can wrap more than one error by providing an Unwrap method that returns a []error.
		// https://tip.golang.org/doc/go1.20#errors

		v := runtime.Version()
		switch {
		case
			strings.HasPrefix(v, "go1.14."),
			strings.HasPrefix(v, "go1.15."),
			strings.HasPrefix(v, "go1.16."),
			strings.HasPrefix(v, "go1.17."),
			strings.HasPrefix(v, "go1.18."),
			strings.HasPrefix(v, "go1.19."):
			t.SkipNow()
		}
		t.Run("errors.Is", func(t *testing.T) {
			err := errors.PrefixedGroup(
				"my group: ",
				errors.PrefixedGroup("some errors: ", io.EOF, io.ErrNoProgress),
				io.ErrUnexpectedEOF,
			)
			assert.True(t, pkgErrors.Is(err, io.ErrNoProgress))
			assert.False(t, pkgErrors.Is(err, io.ErrClosedPipe))
		})
	})
}

func TestCollection(t *testing.T) {
	assert.Nil(t, errors.Collection(nil))
}

func TestNew(t *testing.T) {
	assert.EqualError(t, errors.New("my error"), "my error")
}
