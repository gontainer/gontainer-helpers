package grouperror_test

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/gontainer/gontainer-helpers/v2/grouperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrefixedGroup(t *testing.T) {
	t.Parallel()

	t.Run("Empty", func(t *testing.T) {
		assert.NoError(t, grouperror.Join(nil, nil))
	})

	t.Run("One-error collection", func(t *testing.T) {
		errs := grouperror.Collection(io.EOF)
		assert.Equal(t, []error{io.EOF}, errs)
	})

	t.Run("Errors", func(t *testing.T) {
		fieldsErr := grouperror.Join(
			fmt.Errorf("invalid value of `name`"),
			fmt.Errorf("invalid value of `age`"),
		)

		personErr := grouperror.Prefix(
			"Person: ",
			fieldsErr,
			fmt.Errorf("the given ID does not exist"),
		)
		validationErr := grouperror.Prefix(
			"Validation: ",
			personErr,
			fmt.Errorf("unexpected error"),
		)

		errs := grouperror.Collection(validationErr)
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
	assert.Nil(t, grouperror.Collection(nil))
}

func Test_groupError_Unwrap(t *testing.T) {
	const wrongFileName = "file does not exist"

	getPathError := func() error {
		_, err := os.Open(wrongFileName)
		return err
	}

	err := grouperror.Prefix(
		"my group: ",
		grouperror.Prefix("some errors: ", io.EOF, io.ErrNoProgress),
		io.ErrUnexpectedEOF,
		getPathError(),
	)

	err = grouperror.Prefix("errors: ", err)

	t.Run("errors.Is", func(t *testing.T) {
		for _, target := range []error{io.EOF, io.ErrNoProgress, io.ErrUnexpectedEOF} {
			assert.True(t, errors.Is(err, target))
		}
		assert.False(t, errors.Is(err, io.ErrClosedPipe))
	})

	t.Run("errors.As", func(t *testing.T) {
		t.Run("*os.PathError", func(t *testing.T) {
			var target *os.PathError
			if assert.True(t, errors.As(err, &target)) {
				assert.Equal(t, wrongFileName, target.Path)
			}
		})
		t.Run("*net.AddrError", func(t *testing.T) {
			t.Run("false", func(t *testing.T) {
				var target *net.AddrError
				assert.False(t, errors.As(err, &target))
			})

			t.Run("true", func(t *testing.T) {
				ip := net.IP{1, 2, 3}
				_, addrErr := ip.MarshalText() // address 010203: invalid IP address
				var target *net.AddrError
				assert.Nil(t, target)
				assert.True(
					t,
					errors.As(grouperror.Join(err, addrErr), &target),
				)
				assert.NotNil(t, target)
				assert.EqualError(t, target, "address 010203: invalid IP address")
			})
		})
	})
}
