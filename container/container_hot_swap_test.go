package container

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testContainer struct {
	*container
}

func TestNewContainer_hotSwap(t *testing.T) {
	c := testContainer{NewContainer()}
	s := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ContextWithContainer(ctx, c)
	triggered := false
	c.hotSwap(func(MutableContainer) {
		triggered = true
	})
	assert.True(t, triggered)
	assert.GreaterOrEqual(t, time.Since(s), time.Second)
}
