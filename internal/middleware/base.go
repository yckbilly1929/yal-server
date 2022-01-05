package middleware

import (
	"net/http"
)

// may add recoverer or logger at Base
func Base() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := NewWrapResponseWriter(w)

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
