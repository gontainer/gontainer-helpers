package container

import (
	"net/http"
)

// HTTPHandlerWithContainer creates a new HTTP handler that automatically binds contexts with the container.
func HTTPHandlerWithContainer(handler http.Handler, container Root) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := ContextWithContainer(r.Context(), container)
		r = r.Clone(ctx)
		handler.ServeHTTP(w, r)
	})
}
