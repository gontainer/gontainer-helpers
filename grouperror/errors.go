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

package grouperror

import (
	"errors"
	"fmt"
	"strings"
)

// Join joins provided errors. It ignores nil-values.
// It returns nil, when there are no errors given.
func Join(errs ...error) error {
	return Prefix("", errs...)
}

// Prefix joins errors the same way as [Join], and adds a prefix to the group.
func Prefix(prefix string, errs ...error) error {
	n := 0
	for _, err := range errs {
		if err != nil {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	filtered := make([]error, 0, n)
	for _, err := range errs {
		if err != nil {
			filtered = append(filtered, err)
		}
	}

	return &groupError{
		prefix: prefix,
		errors: filtered,
	}
}

type groupError struct {
	prefix string
	errors []error
}

func (g *groupError) Error() string {
	c := g.Collection()
	s := make([]string, 0, len(c))
	for _, err := range c {
		s = append(s, err.Error())
	}
	return strings.Join(s, "\n")
}

func (g *groupError) Unwrap() []error {
	return g.Collection()
}

func (g *groupError) Collection() []error {
	var errs []error
	for _, err := range g.errors {
		if subGroupErr, ok := err.(interface{ Collection() []error }); ok {
			for _, nErr := range subGroupErr.Collection() {
				errs = append(errs, fmt.Errorf("%s%w", g.prefix, nErr))
			}
			continue
		}
		errs = append(errs, fmt.Errorf("%s%w", g.prefix, err))
	}
	return errs
}

// Is provides support for `errors.Is` in older versions of Go (<1.20)
//
// https://tip.golang.org/doc/go1.20#errors
func (g *groupError) Is(target error) bool {
	for _, err := range g.errors {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// As provides support for `errors.As` in older versions of Go (<1.20)
//
// https://tip.golang.org/doc/go1.20#errors
func (g *groupError) As(target any) bool {
	for _, err := range g.errors {
		if errors.As(err, target) {
			return true
		}
	}
	return false
}

func Collection(err error) []error {
	if err == nil {
		return nil
	}
	if collection, ok := err.(interface{ Collection() []error }); ok {
		return collection.Collection()
	}
	return []error{err}
}
