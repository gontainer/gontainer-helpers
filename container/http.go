package container

import (
	"net/http"
)

// TODO is it a good idea to add it to this package?
func HTTPHandlerWithContainer(handler http.Handler, container Self) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := ContextWithContainer(r.Context(), container)
		r = r.Clone(ctx)
		handler.ServeHTTP(w, r)
	})
}
