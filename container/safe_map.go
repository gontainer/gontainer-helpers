package container

import (
	"sync"
)

// safeMap provides the interface for concurrent-safe operations over a map.
type safeMap struct {
	data   map[string]interface{}
	locker rwlocker
}

func newSafeMap() *safeMap {
	return &safeMap{
		data:   make(map[string]interface{}),
		locker: &sync.RWMutex{},
	}
}

func (s safeMap) set(id string, v interface{}) {
	s.locker.Lock()
	defer s.locker.Unlock()

	s.data[id] = v
}

func (s safeMap) get(id string) (value interface{}, exists bool) {
	s.locker.RLock()
	defer s.locker.RUnlock()

	v, ok := s.data[id]
	return v, ok
}

func (s safeMap) delete(id string) {
	s.locker.Lock()
	defer s.locker.Unlock()

	delete(s.data, id)
}
