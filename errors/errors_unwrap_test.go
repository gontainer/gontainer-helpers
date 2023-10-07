//go:build go1.20
// +build go1.20

package errors_test

import (
	pkgErrors "errors"
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/gontainer/gontainer-helpers/errors"
	"github.com/stretchr/testify/assert"
)

func Test_groupError_Unwrap(t *testing.T) {
	// https://tip.golang.org/doc/go1.20#errors

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

	t.Run("errors.Is", func(t *testing.T) {
		for _, target := range []error{io.EOF, io.ErrNoProgress, io.ErrUnexpectedEOF} {
			assert.True(t, pkgErrors.Is(err, target))
		}
		assert.False(t, pkgErrors.Is(err, io.ErrClosedPipe))
	})

	t.Run("errors.As", func(t *testing.T) {
		var target *fs.PathError
		if assert.True(t, pkgErrors.As(err, &target)) {
			assert.Equal(t, wrongFileName, target.Path)
		}
	})
}
