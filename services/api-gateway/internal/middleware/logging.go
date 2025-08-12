package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/ekyc-backend/pkg/logger"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Logging middleware logs HTTP request/response details
func Logging(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start)

			// Extract context values
			ctx := r.Context()
			requestID := GetRequestID(ctx)
			correlationID := GetCorrelationID(ctx)
			userID := GetUserID(ctx)
			sessionID := GetSessionID(ctx)

			// Get trace context
			var traceID, spanID string
			if span := trace.SpanFromContext(ctx); span != nil {
				spanCtx := span.SpanContext()
				traceID = spanCtx.TraceID().String()
				spanID = spanCtx.SpanID().String()
			}

			// Log request details
			log.WithContext(ctx).WithTraceID(traceID).
				WithSpanID(spanID).
				WithCorrelationID(correlationID).
				WithSessionID(sessionID).
				Info("HTTP Request",
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.Int("status", wrapped.statusCode),
					zap.Int64("duration_ms", duration.Milliseconds()),
					zap.String("remote_ip", getRemoteIP(r)),
					zap.String("user_agent", r.UserAgent()),
					zap.String("request_id", requestID),
					zap.String("user_id", userID),
				)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// getRemoteIP extracts the real IP address from the request
func getRemoteIP(r *http.Request) string {
	// Check for forwarded headers
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	// Fallback to remote address
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}

	return "unknown"
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return ""
}

// GetSessionID extracts session ID from context
func GetSessionID(ctx context.Context) string {
	if sessionID, ok := ctx.Value("session_id").(string); ok {
		return sessionID
	}
	return ""
}
