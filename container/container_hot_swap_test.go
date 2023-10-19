package container

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewContainer_hotSwap(t *testing.T) {
	c := NewContainer()
	s := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ContextWithContainer(ctx, c)
	triggered := false
	c.hotSwap(func() {
		triggered = true
	})
	assert.True(t, triggered)
	assert.GreaterOrEqual(t, time.Since(s), time.Second)
}
