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

package grouperror_test

import (
	"errors"
	"fmt"

	"github.com/gontainer/gontainer-helpers/v3/grouperror"
)

func Example() {
	err := grouperror.Prefix("my group: ", fmt.Errorf("error1"), nil, fmt.Errorf("error2"))
	for _, x := range grouperror.Collection(err) {
		fmt.Println(x)
	}
	// Output:
	// my group: error1
	// my group: error2
}

func ExamplePrefix() {
	err := grouperror.Prefix(
		"validation: ",
		errors.New("invalid name"),
		nil, // nil-errors are being ignored
		nil,
		errors.New("invalid age"),
	)

	err = grouperror.Prefix(
		"could not create new user: ",
		errors.New("unexpected error"),
		err,
	)

	err = grouperror.Prefix("operation failed: ", err)

	fmt.Println(err.Error())
	fmt.Println()

	for i, cErr := range grouperror.Collection(err) {
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
