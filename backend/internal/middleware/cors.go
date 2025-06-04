package middleware

import (
	"net/http"
)

// CORS wraps any handler to add the proper headers.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow your frontend origin
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		// Allow the methods your frontend might use
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// Allow the headers (e.g. Content-Type and Authorization for JWT)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// If this is a preflight request, return 200 directly
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
