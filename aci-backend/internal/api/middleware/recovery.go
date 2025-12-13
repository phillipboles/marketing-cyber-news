package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog/log"
)

// Recoverer is a middleware that recovers from panics and logs the stack trace
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestID(r.Context())

				// Log the panic with stack trace
				log.Error().
					Str("request_id", requestID).
					Interface("panic", err).
					Bytes("stack", debug.Stack()).
					Msg("Panic recovered")

				// Return 500 Internal Server Error
				w.Header().Set("Content-Type", "application/json")
				if requestID != "" {
					w.Header().Set("X-Request-ID", requestID)
				}
				w.WriteHeader(http.StatusInternalServerError)

				// Write error response
				response := `{"error":{"code":"INTERNAL_ERROR","message":"An unexpected error occurred"`
				if requestID != "" {
					response += `,"request_id":"` + requestID + `"`
				}
				response += `}}`
				w.Write([]byte(response))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
