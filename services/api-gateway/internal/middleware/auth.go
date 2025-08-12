package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ekyc-backend/pkg/logger"
	"github.com/ekyc-backend/services/api-gateway/internal/security"
	"go.uber.org/zap"
)

const (
	// AuthorizationHeader is the header name for authorization
	AuthorizationHeader = "Authorization"
	// BearerPrefix is the prefix for Bearer tokens
	BearerPrefix = "Bearer "
	// UserIDContextKey is the context key for user ID
	UserIDContextKey = "user_id"
	// UserRolesContextKey is the context key for user roles
	UserRolesContextKey = "user_roles"
)

// Auth middleware validates JWT tokens and extracts user information
func Auth(jwtManager *security.JWTManager, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract authorization header
			authHeader := r.Header.Get(AuthorizationHeader)
			if authHeader == "" {
				log.WithContext(ctx).Warn("Missing authorization header",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
				)
				writeUnauthorizedResponse(w, "Missing authorization header")
				return
			}

			// Check Bearer prefix
			if !strings.HasPrefix(authHeader, BearerPrefix) {
				log.WithContext(ctx).Warn("Invalid authorization header format",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
				)
				writeUnauthorizedResponse(w, "Invalid authorization header format")
				return
			}

			// Extract token
			token := strings.TrimPrefix(authHeader, BearerPrefix)
			if token == "" {
				log.WithContext(ctx).Warn("Empty token",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
				)
				writeUnauthorizedResponse(w, "Empty token")
				return
			}

			// Validate token
			claims, err := jwtManager.ValidateToken(token)
			if err != nil {
				log.WithContext(ctx).Warn("Invalid token",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
					zap.Error(err),
				)
				writeUnauthorizedResponse(w, "Invalid token")
				return
			}

			// Add user information to context
			ctx = context.WithValue(ctx, UserIDContextKey, claims.UserID)
			ctx = context.WithValue(ctx, UserRolesContextKey, claims.Roles)
			r = r.WithContext(ctx)

			// Proceed with request
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole middleware ensures the user has a specific role
func RequireRole(requiredRole string, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userRoles := GetUserRoles(ctx)

			// Check if user has required role
			hasRole := false
			for _, role := range userRoles {
				if role == requiredRole {
					hasRole = true
					break
				}
			}

			if !hasRole {
				log.WithContext(ctx).Warn("Insufficient permissions",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
					zap.String("required_role", requiredRole),
					zap.Strings("user_roles", userRoles),
				)
				writeForbiddenResponse(w, "Insufficient permissions")
				return
			}

			// Proceed with request
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole middleware ensures the user has at least one of the required roles
func RequireAnyRole(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userRoles := GetUserRoles(ctx)

			// Check if user has any of the required roles
			hasRole := false
			for _, requiredRole := range requiredRoles {
				for _, userRole := range userRoles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				writeForbiddenResponse(w, "Insufficient permissions")
				return
			}

			// Proceed with request
			next.ServeHTTP(w, r)
		})
	}
}

// writeUnauthorizedResponse writes a 401 Unauthorized response
func writeUnauthorizedResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "UNAUTHORIZED",
			"message": message,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// writeForbiddenResponse writes a 403 Forbidden response
func writeForbiddenResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)

	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    "FORBIDDEN",
			"message": message,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// GetUserRoles extracts user roles from context
func GetUserRoles(ctx context.Context) []string {
	if userRoles, ok := ctx.Value(UserRolesContextKey).([]string); ok {
		return userRoles
	}
	return []string{}
}
