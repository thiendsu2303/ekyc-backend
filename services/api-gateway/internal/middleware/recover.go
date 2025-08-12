package middleware

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"github.com/ekyc-backend/pkg/logger"
	"go.uber.org/zap"
)

// Recover middleware recovers from panics and returns a 500 error
func Recover(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic
					log.WithContext(r.Context()).Error("Panic recovered",
						zap.Any("error", err),
						zap.String("stack", string(debug.Stack())),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.String("remote_addr", r.RemoteAddr),
					)

					// Return 500 Internal Server Error
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)

					response := map[string]interface{}{
						"error": map[string]interface{}{
							"code":    "INTERNAL_ERROR",
							"message": "Internal server error",
							"details": "A panic occurred and was recovered",
						},
					}

					// Add correlation ID if available
					if correlationID := GetCorrelationID(r.Context()); correlationID != "" {
						response["error"].(map[string]interface{})["correlationId"] = correlationID
					}

					json.NewEncoder(w).Encode(response)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
