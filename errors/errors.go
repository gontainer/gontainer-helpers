package errors

import (
	"errors"

	"github.com/gontainer/gontainer-helpers/grouperror"
)

var (
	// Deprecated: use errors.New.
	New = errors.New

	// Deprecated: use grouperror.Join.
	Group = grouperror.Join

	// Deprecated: use grouperror.Prefix.
	PrefixedGroup = grouperror.Prefix

	// Deprecated: use grouperror.Collection.
	Collection = grouperror.Collection
)
