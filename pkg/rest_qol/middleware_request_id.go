package rest_qol

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RequestIDMiddleware ensures each request has an X-Request-Id header.
// The prefix is applied only when the incoming request does not already include one.
func RequestIDMiddleware(prefix string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			EnsureRequestID(w, r, prefix)
			next.ServeHTTP(w, r)
		})
	}
}
