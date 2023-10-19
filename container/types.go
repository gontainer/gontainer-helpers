package container

type rwlocker interface {
	RLock()
	RUnlock()
	Lock()
	Unlock()
}

type keyValue interface {
	set(id string, v any)
	get(id string) (result any, exists bool)
	delete(id string)
}
