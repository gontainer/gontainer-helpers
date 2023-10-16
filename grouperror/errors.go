package grouperror

import (
	"errors"
	"fmt"
	"strings"
)

// Join joins provided errors. It ignores nil nil-values.
// It may return nil, when there are no errors given.
func Join(errs ...error) error {
	return Prefix("", errs...)
}

// Prefix joins error the same way as Join, and adds a prefix to the group.
func Prefix(prefix string, errs ...error) error {
	filtered := make([]error, 0, len(errs))
	for _, err := range errs {
		if err != nil {
			filtered = append(filtered, err)
		}
	}
	if len(filtered) == 0 {
		return nil
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
func (g *groupError) As(target interface{}) bool {
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
