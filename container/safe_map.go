// Copyright (c) 2023 Bartłomiej Krukowski
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is furnished
// to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package container

import (
	"sync"
)

// safeMap provides the interface for concurrent-safe operations over a map.
type safeMap struct {
	data   map[string]any
	locker rwlocker
}

func newSafeMap() *safeMap {
	return &safeMap{
		data:   make(map[string]any),
		locker: &sync.RWMutex{},
	}
}

func (s *safeMap) set(id string, v any) {
	s.locker.Lock()
	defer s.locker.Unlock()

	s.data[id] = v
}

func (s *safeMap) get(id string) (value any, exists bool) {
	s.locker.RLock()
	defer s.locker.RUnlock()

	v, ok := s.data[id]
	return v, ok
}

func (s *safeMap) delete(id string) {
	s.locker.Lock()
	defer s.locker.Unlock()

	delete(s.data, id)
}
