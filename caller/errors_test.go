package caller

import (
	"errors"
	"testing"

	"github.com/gontainer/gontainer-helpers/v3/grouperror"
	errAssert "github.com/gontainer/gontainer-helpers/v3/grouperror/assert"
)

func TestProviderError_Collection(t *testing.T) {
	err := newProviderError(grouperror.Prefix("prefix: ", errors.New("error 1"), errors.New("error 2")))
	expected := []string{
		`prefix: error 1`,
		`prefix: error 2`,
	}
	errAssert.EqualErrorGroup(t, err, expected)
}
