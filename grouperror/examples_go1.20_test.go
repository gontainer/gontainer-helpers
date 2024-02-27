// Copyright (c) 2023–present Bartłomiej Krukowski
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

//go:build go1.20
// +build go1.20

package grouperror_test

import (
	"errors"
	"fmt"
)

func ExamplePrefix_stdlib() {
	// [errors.Join] has been introduced in go 1.20
	// https://tip.golang.org/doc/go1.20#errors
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
