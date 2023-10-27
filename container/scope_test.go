package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_scope_String(t *testing.T) {
	assert.Equal(t, "scopeNonShared", scopeNonShared.String())
	assert.Equal(t, "unknown", (scopeNonShared + 1).String())
}
