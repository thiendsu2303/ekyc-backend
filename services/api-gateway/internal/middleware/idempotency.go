package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/ekyc-backend/pkg/logger"
	"github.com/ekyc-backend/pkg/storage"
	"go.uber.org/zap"
)

const (
	// IdempotencyKeyHeader is the header name for idempotency key
	IdempotencyKeyHeader = "Idempotency-Key"
	// IdempotencyTTL is the TTL for idempotency cache (15 minutes)
	IdempotencyTTL = 15 * time.Minute
)

// Idempotency middleware ensures idempotent operations
func Idempotency(redisClient *storage.Redis, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only apply to POST requests
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			// Check if idempotency key is provided
			idempotencyKey := r.Header.Get(IdempotencyKeyHeader)
			if idempotencyKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			route := r.URL.Path

			// Create cache key
			cacheKey := "idempotency:" + route + ":" + idempotencyKey

			// Check if we have a cached response
			cachedResponse, err := redisClient.Get(ctx, cacheKey)
			if err == nil && cachedResponse != "" {
				// Return cached response
				log.WithContext(ctx).Info("Idempotency cache hit",
					zap.String("route", route),
					zap.String("key", idempotencyKey),
				)

				var response map[string]interface{}
				if err := json.Unmarshal([]byte(cachedResponse), &response); err == nil {
					// Set response headers
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("X-Idempotency-Cache", "hit")

					// Write cached response
					json.NewEncoder(w).Encode(response)
					return
				}
			}

			// Create response writer wrapper to capture response
			wrapped := &idempotencyResponseWriter{
				ResponseWriter: w,
				body:           make([]byte, 0),
				statusCode:     http.StatusOK,
				headers:        make(http.Header),
			}

			// Process request
			next.ServeHTTP(wrapped, r)

			// Cache successful responses (2xx status codes)
			if wrapped.statusCode >= 200 && wrapped.statusCode < 300 {
				response := map[string]interface{}{
					"status":  wrapped.statusCode,
					"headers": wrapped.headers,
					"body":    string(wrapped.body),
				}

				responseJSON, err := json.Marshal(response)
				if err == nil {
					// Cache the response
					err = redisClient.Set(ctx, cacheKey, string(responseJSON), IdempotencyTTL)
					if err != nil {
						log.WithContext(ctx).Error("Failed to cache idempotency response",
							zap.String("route", route),
							zap.String("key", idempotencyKey),
							zap.Error(err),
						)
					}
				}
			}
		})
	}
}

// idempotencyResponseWriter wraps http.ResponseWriter to capture response
type idempotencyResponseWriter struct {
	http.ResponseWriter
	body       []byte
	statusCode int
	headers    http.Header
}

func (irw *idempotencyResponseWriter) WriteHeader(code int) {
	irw.statusCode = code
	irw.ResponseWriter.WriteHeader(code)
}

func (irw *idempotencyResponseWriter) Write(b []byte) (int, error) {
	irw.body = append(irw.body, b...)
	return irw.ResponseWriter.Write(b)
}

func (irw *idempotencyResponseWriter) Header() http.Header {
	if irw.headers == nil {
		irw.headers = make(http.Header)
	}
	return irw.headers
}

// generateIdempotencyKey generates a hash from request body for idempotency
func generateIdempotencyKey(body []byte) string {
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

// readRequestBody reads and returns the request body
func readRequestBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}
