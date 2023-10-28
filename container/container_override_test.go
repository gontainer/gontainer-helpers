package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_overrideService(t *testing.T) {
	t.Run("Invalid scope", func(t *testing.T) {
		defer func() {
			assert.Equal(t, `overrideService("service"): invalid scope "unknown"`, recover())
		}()

		s := NewService()
		s.SetValue(nil)
		s.scope = scopeNonShared + 1

		c := New()
		c.OverrideService("service", s)
	})
}
