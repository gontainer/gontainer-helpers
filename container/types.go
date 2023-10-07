package container

type rwlocker interface {
	RLock()
	RUnlock()
	Lock()
	Unlock()
}

type keyValue interface {
	set(id string, v interface{})
	get(id string) (result interface{}, exists bool)
	delete(id string)
}
