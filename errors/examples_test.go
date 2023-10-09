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
		errors.New("unexpected error"),
		err,
	)

	err = errors.PrefixedGroup("operation failed: ", err)

	fmt.Println(err.Error())
	fmt.Println()

	for i, cErr := range errors.Collection(err) {
		fmt.Printf("%d. %s\n", i+1, cErr.Error())
	}

	// Output:
	// operation failed: could not create new user: unexpected error
	// operation failed: could not create new user: validation: invalid name
	// operation failed: could not create new user: validation: invalid age
	//
	// 1. operation failed: could not create new user: unexpected error
	// 2. operation failed: could not create new user: validation: invalid name
	// 3. operation failed: could not create new user: validation: invalid age
}
