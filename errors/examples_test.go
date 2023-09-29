package errors_test

import (
	"fmt"

	"github.com/gontainer/gontainer-helpers/errors"
)

func ExamplePrefixedGroup() {
	err := errors.PrefixedGroup(
		"validation: ",
		errors.New("invalid name"),
		nil, // nil-errors are being ignored
		nil,
		errors.New("invalid age"),
	)

	err = errors.PrefixedGroup(
		"could not create new user: ",
		err,
		errors.New("unexpected error"),
	)

	err = errors.PrefixedGroup("operation failed: ", err)

	for i, cErr := range errors.Collection(err) {
		fmt.Printf("%d. %s\n", i+1, cErr.Error())
	}

	// Output:
	// 1. operation failed: could not create new user: validation: invalid name
	// 2. operation failed: could not create new user: validation: invalid age
	// 3. operation failed: could not create new user: unexpected error
}
