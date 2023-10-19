//go:build go1.21
// +build go1.21

package groupcontext

import (
	"context"
)

func (g *groupContext) Add(ctx context.Context) {
	g.assertValidContext(ctx)
	g.waitGroup.Add(1)
	context.AfterFunc(ctx, g.waitGroup.Done)
}
