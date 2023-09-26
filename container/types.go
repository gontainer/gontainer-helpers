package container

type rwlocker interface {
	RLock()
	RUnlock()
	Lock()
	Unlock()
}
