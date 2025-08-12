package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"
	// RequestIDContextKey is the context key for request ID
	RequestIDContextKey = "request_id"
)

// RequestID middleware adds a unique request ID to each request
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if request ID is already provided
			requestID := r.Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Add request ID to response headers
			w.Header().Set(RequestIDHeader, requestID)

			// Add request ID to request context
			ctx := r.Context()
			ctx = context.WithValue(ctx, RequestIDContextKey, requestID)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDContextKey).(string); ok {
		return requestID
	}
	return ""
}
