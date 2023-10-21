package graph

import (
	"fmt"
	"strings"

	"github.com/gontainer/gontainer-helpers/v2/grouperror"
)

func CircularDepsToError(circularDeps [][]Dependency) error {
	var errs []error

	for _, cycle := range circularDeps {
		ids := make([]string, 0, len(cycle))
		for _, node := range cycle {
			ids = append(ids, node.Pretty)
		}
		errs = append(errs, fmt.Errorf("%s", strings.Join(ids, " -> ")))
	}

	return grouperror.Join(errs...)
}
