package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	// CorrelationIDHeader is the header name for correlation ID
	CorrelationIDHeader = "X-Correlation-ID"
	// CorrelationIDContextKey is the context key for correlation ID
	CorrelationIDContextKey = "correlation_id"
)

// CorrelationID middleware adds a correlation ID to each request
func CorrelationID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if correlation ID is already provided
			correlationID := r.Header.Get(CorrelationIDHeader)
			if correlationID == "" {
				correlationID = uuid.New().String()
			}

			// Add correlation ID to response headers
			w.Header().Set(CorrelationIDHeader, correlationID)

			// Add correlation ID to request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, CorrelationIDContextKey, correlationID)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// GetCorrelationID extracts correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value(CorrelationIDContextKey).(string); ok {
		return correlationID
	}
	return ""
}
