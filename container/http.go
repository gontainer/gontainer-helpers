package container

import (
	"net/http"
)

// TODO is it a good idea to add it to this package?
func HTTPHandlerWithContainer(h http.Handler, c Self) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := ContextWithContainer(r.Context(), c)
		r = r.Clone(ctx)
		h.ServeHTTP(w, r)
	})
}
