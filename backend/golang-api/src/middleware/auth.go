package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"
)

func OpenClawAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		apiKey := os.Getenv("OPENCLAW_API_KEY")
		if apiKey == "" {
			log.Println("WARNING: OPENCLAW_API_KEY not set, skipping auth")
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"success":false,"message":"Missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] != apiKey {
			http.Error(w, `{"success":false,"message":"Invalid API key"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
