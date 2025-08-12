package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/ekyc-backend/pkg/logger"
	"github.com/ekyc-backend/pkg/storage"
	"go.uber.org/zap"
)

// RateLimit middleware implements token bucket rate limiting
func RateLimit(redisClient *storage.Redis, rps, burst int, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ip := getRemoteIP(r)
			route := r.URL.Path

			// Create rate limit key
			key := "rate_limit:" + ip + ":" + route

			// Check if rate limit exceeded
			allowed, err := redisClient.CheckRateLimit(ctx, key, rps, time.Duration(burst)*time.Second)
			if err != nil {
				log.WithContext(ctx).Error("Rate limit check failed",
					zap.String("ip", ip),
					zap.String("route", route),
					zap.Error(err),
				)
				// On error, allow the request to proceed
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				// Rate limit exceeded
				log.WithContext(ctx).Warn("Rate limit exceeded",
					zap.String("ip", ip),
					zap.String("route", route),
					zap.Int("rps", rps),
					zap.Int("burst", burst),
				)

				// Return 429 Too Many Requests
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", strconv.Itoa(60)) // Retry after 1 minute
				w.WriteHeader(http.StatusTooManyRequests)

				response := map[string]interface{}{
					"error": map[string]interface{}{
						"code":    "RATE_LIMIT_EXCEEDED",
						"message": "Rate limit exceeded",
						"details": "Too many requests from this IP address",
					},
				}

				// Add correlation ID if available
				if correlationID := GetCorrelationID(ctx); correlationID != "" {
					response["error"].(map[string]interface{})["correlationId"] = correlationID
				}

				json.NewEncoder(w).Encode(response)
				return
			}

			// Rate limit check passed, proceed with request
			next.ServeHTTP(w, r)
		})
	}
}
