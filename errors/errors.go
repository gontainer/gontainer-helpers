package errors

import (
	"errors"
	"fmt"
	"strings"
)

// New is an alias to errors.New
func New(s string) error {
	return errors.New(s)
}

// Group joins provided errors. It ignores nil nil-values.
// It may return nil, when there are not errors given.
func Group(errs ...error) error {
	return PrefixedGroup("", errs...)
}

// PrefixedGroup joins error the same way as Group, and adds a prefix to the group.
func PrefixedGroup(prefix string, errs ...error) error {
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
		errs:   filtered,
	}
}

type groupError struct {
	prefix string
	errs   []error
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
	for _, err := range g.errs {
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

// Is supports errors.Is for go < 1.20
//
// https://tip.golang.org/doc/go1.20#errors
func (g *groupError) Is(target error) bool {
	for _, err := range g.Unwrap() {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

// As supports errors.As for go < 1.20
//
// https://tip.golang.org/doc/go1.20#errors
func (g *groupError) As(target interface{}) bool {
	for _, err := range g.Unwrap() {
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
