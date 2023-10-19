package container

// TODO
func (c *container) hotSwap(fn func()) {
	c.groupContext.Wait()
	fn()
}
