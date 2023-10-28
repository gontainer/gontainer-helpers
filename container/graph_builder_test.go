package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_graphBuilder_resolveScope(t *testing.T) {
	t.Run("Panic", func(t *testing.T) {
		defer func() {
			assert.Equal(t, `scope for "unknownService" does not exist in cache`, recover())
		}()
		c := New()
		c.graphBuilder.resolveScope("unknownService")
	})
}
