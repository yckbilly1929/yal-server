package middleware

import (
	"net/http"
)

func Intercept404(goodHandler, badHandler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bw, ok := w.(WrapResponseWriter)
		if !ok {
			goodHandler.ServeHTTP(w, r)
			return
		}

		goodHandler.ServeHTTP(bw, r)
		if bw.Status() == http.StatusNotFound {
			badHandler.ServeHTTP(bw.Unwrap(), r)
		}
	}
}
