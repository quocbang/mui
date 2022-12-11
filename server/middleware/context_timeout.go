package middleware

import (
	"context"
	"net/http"
	"time"
)

// ContextTimeoutMiddleware handlers server timeout duration.
func ContextTimeoutMiddleware(next http.Handler, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if timeout <= 0 {
			next.ServeHTTP(w, r)
			return
		}

		newContext, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		next.ServeHTTP(w, r.WithContext(newContext))
	})
}
