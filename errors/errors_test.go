package errors_test

import (
	stdErrors "errors"
	"fmt"
	"io"
	"net"
	"os"
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
}

func TestCollection(t *testing.T) {
	assert.Nil(t, errors.Collection(nil))
}

func TestNew(t *testing.T) {
	assert.EqualError(t, errors.New("my error"), "my error")
}

func Test_groupError_Unwrap(t *testing.T) {
	const wrongFileName = "file does not exist"

	getPathError := func() error {
		_, err := os.Open(wrongFileName)
		return err
	}

	err := errors.PrefixedGroup(
		"my group: ",
		errors.PrefixedGroup("some errors: ", io.EOF, io.ErrNoProgress),
		io.ErrUnexpectedEOF,
		getPathError(),
	)

	err = errors.PrefixedGroup("errors: ", err)

	t.Run("errors.Is", func(t *testing.T) {
		for _, target := range []error{io.EOF, io.ErrNoProgress, io.ErrUnexpectedEOF} {
			assert.True(t, stdErrors.Is(err, target))
		}
		assert.False(t, stdErrors.Is(err, io.ErrClosedPipe))
	})

	t.Run("errors.As", func(t *testing.T) {
		t.Run("*os.PathError", func(t *testing.T) {
			var target *os.PathError
			if assert.True(t, stdErrors.As(err, &target)) {
				assert.Equal(t, wrongFileName, target.Path)
			}
		})
		t.Run("*net.AddrError", func(t *testing.T) {
			t.Run("false", func(t *testing.T) {
				var target *net.AddrError
				assert.False(t, stdErrors.As(err, &target))
			})

			t.Run("true", func(t *testing.T) {
				ip := make(net.IP, 1)
				_, addrErr := ip.MarshalText() // address 00: invalid IP address
				var target *net.AddrError
				assert.True(
					t,
					stdErrors.As(errors.Group(err, addrErr), &target),
				)
			})
		})
	})
}
