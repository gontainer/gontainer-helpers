package graph

import (
	"github.com/gontainer/gontainer-helpers/v2/container/graph/internal/graph"
)

var (
	// Deprecated: do not use it, exported for gontainer/gontainer only.
	New = graph.New

	// Deprecated: do not use it, exported for gontainer/gontainer only.
	CircularDepsToError = graph.CircularDepsToError
)

type (
	// Deprecated: do not use it, exported for gontainer/gontainer only.
	Dependency = graph.Dependency
)
