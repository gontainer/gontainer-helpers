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

package http

import (
	"fmt"
	"net/http"

	"github.com/gontainer/gontainer-helpers/v3/container"
)

// HandlerWithContainer creates a new HTTP handler that automatically binds contexts with the container.
func HandlerWithContainer(container_ container.Root, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := container.ContextWithContainer(r.Context(), container_)
		r = r.Clone(ctx)
		handler.ServeHTTP(w, r)
	})
}

// ServeMux extends [http.ServeMux].
// It automatically binds the contexts of registered handlers with the given container
// and lets us build handlers in real-time (for the contextual scopes).
type ServeMux struct {
	*http.ServeMux
	container container.Root
}

// NewServeMux allocates and returns a new ServeMux.
func NewServeMux(container container.Root) *ServeMux {
	return &ServeMux{
		ServeMux:  http.NewServeMux(),
		container: container,
	}
}

// Handle wraps the provided handler by [HandlerWithContainer] and registers the handler for the given pattern.
func (s *ServeMux) Handle(pattern string, handler http.Handler) {
	s.ServeMux.Handle(pattern, HandlerWithContainer(s.container, handler))
}

/*
HandleDynamic registers the handler for the given pattern.
Since the handler is fetched from the [*container.Container], it allows for building the handler in runtime.
It is useful when our handler has contextual dependencies.
*/
func (s *ServeMux) HandleDynamic(pattern string, handlerID string) {
	s.Handle(
		pattern,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tmp, err := s.container.Root().Get(handlerID)
			if err != nil {
				panic(fmt.Sprintf(
					"HandleDynamic %+q: container returned error: %s",
					pattern,
					err.Error(),
				))
			}
			h, ok := tmp.(http.Handler)
			if !ok {
				panic(fmt.Sprintf(
					"HandleDynamic %+q: service %+q does not implement http.Handler",
					pattern,
					handlerID,
				))
			}
			h.ServeHTTP(w, r)
		}),
	)
}
