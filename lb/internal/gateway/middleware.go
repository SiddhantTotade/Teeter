package gateway

import (
	"net/http"
)

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		if token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		}

		next.ServeHTTP(w, r)
	})
}
