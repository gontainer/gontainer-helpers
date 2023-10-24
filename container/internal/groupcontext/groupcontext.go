package groupcontext

import (
	"context"
	"sync"
)

type groupContext struct {
	waitGroup *sync.WaitGroup
}

func New() *groupContext {
	return &groupContext{
		waitGroup: new(sync.WaitGroup),
	}
}

func (g *groupContext) Wait() {
	g.waitGroup.Wait()
}

func (g *groupContext) assertValidContext(ctx context.Context) {
	if ctx.Done() == nil {
		// https://dave.cheney.net/2014/03/19/channel-axioms
		panic("ctx.Done() == nil: a receive from a nil channel blocks forever")
	}
}
