package middleware

import (
	"net/http"
	"os"
)

func APIKeyAuth(next http.Handler) http.Handler {
	expectedKey := os.Getenv("API_KEY")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "" || key != expectedKey {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// add info to context
		//ctx := context.WithValue(r.Context(), apiKeyCtxKey, "admin")

		next.ServeHTTP(w, r)
	})
}
