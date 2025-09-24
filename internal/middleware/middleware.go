package middleware

import (
	"context"
	"github.com/rs/cors"
	"net/http"
)

// RequestMiddleware add request context middleware
func RequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "httpRequest", r)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CORSHandler returns a middleware that applies CORS settings
func CORSHandler(next http.Handler) http.Handler {
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"}, // Adjust as needed
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	return corsHandler.Handler(next)
}
