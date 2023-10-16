//go:build go1.20
// +build go1.20

package grouperror_test

import (
	"errors"
	"fmt"
)

func ExamplePrefix_stdlib() {
	// we need the following build tags:
	//
	// //go:build go1.20
	// // +build go1.20
	//
	// because errors.Join has been introduced in go 1.20
	err := errors.Join(
		errors.New("invalid name"),
		nil,
		nil,
		errors.New("invalid age"),
	)

	err = fmt.Errorf("validation: %w", err)

	err = errors.Join(
		errors.New("unexpected error"),
		err,
	)

	err = fmt.Errorf("could not create new user: %w", err)

	err = fmt.Errorf("operation failed: %w", err)

	fmt.Println(err.Error())

	// Output:
	// operation failed: could not create new user: unexpected error
	// validation: invalid name
	// invalid age
}
