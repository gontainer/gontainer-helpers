package graph

import (
	"fmt"
	"strings"

	"github.com/gontainer/gontainer-helpers/v2/grouperror"
)

func CircularDepsToError(circularDeps [][]Dependency) error {
	errs := make([]error, len(circularDeps))

	for i, cycle := range circularDeps {
		ids := make([]string, len(cycle))
		for j, node := range cycle {
			ids[j] = node.Pretty
		}
		errs[i] = fmt.Errorf("%s", strings.Join(ids, " -> "))
	}

	return grouperror.Join(errs...)
}
