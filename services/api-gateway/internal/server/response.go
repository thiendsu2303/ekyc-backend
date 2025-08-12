package server

import (
	"encoding/json"
	"net/http"

	"github.com/ekyc-backend/pkg/logger"
	"go.uber.org/zap"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error represents an API error
type Error struct {
	Code          string `json:"code"`
	Message       string `json:"message"`
	Details       string `json:"details,omitempty"`
	CorrelationID string `json:"correlationId,omitempty"`
}

// SuccessResponse sends a successful response
func SuccessResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	response := Response{
		Success: true,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// ErrorResponse sends an error response
func ErrorResponse(w http.ResponseWriter, code, message, details string, statusCode int, correlationID string) {
	response := Response{
		Success: false,
		Error: &Error{
			Code:          code,
			Message:       message,
			Details:       details,
			CorrelationID: correlationID,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// BadRequestResponse sends a 400 Bad Request response
func BadRequestResponse(w http.ResponseWriter, message, details, correlationID string) {
	ErrorResponse(w, "BAD_REQUEST", message, details, http.StatusBadRequest, correlationID)
}

// UnauthorizedResponse sends a 401 Unauthorized response
func UnauthorizedResponse(w http.ResponseWriter, message, correlationID string) {
	ErrorResponse(w, "UNAUTHORIZED", message, "", http.StatusUnauthorized, correlationID)
}

// ForbiddenResponse sends a 403 Forbidden response
func ForbiddenResponse(w http.ResponseWriter, message, correlationID string) {
	ErrorResponse(w, "FORBIDDEN", message, "", http.StatusForbidden, correlationID)
}

// NotFoundResponse sends a 404 Not Found response
func NotFoundResponse(w http.ResponseWriter, message, correlationID string) {
	ErrorResponse(w, "NOT_FOUND", message, "", http.StatusNotFound, correlationID)
}

// ConflictResponse sends a 409 Conflict response
func ConflictResponse(w http.ResponseWriter, message, details, correlationID string) {
	ErrorResponse(w, "CONFLICT", message, details, http.StatusConflict, correlationID)
}

// TooManyRequestsResponse sends a 429 Too Many Requests response
func TooManyRequestsResponse(w http.ResponseWriter, message, correlationID string) {
	ErrorResponse(w, "TOO_MANY_REQUESTS", message, "", http.StatusTooManyRequests, correlationID)
}

// InternalServerErrorResponse sends a 500 Internal Server Error response
func InternalServerErrorResponse(w http.ResponseWriter, message, correlationID string) {
	ErrorResponse(w, "INTERNAL_ERROR", message, "", http.StatusInternalServerError, correlationID)
}

// ServiceUnavailableResponse sends a 503 Service Unavailable response
func ServiceUnavailableResponse(w http.ResponseWriter, message, correlationID string) {
	ErrorResponse(w, "SERVICE_UNAVAILABLE", message, "", http.StatusServiceUnavailable, correlationID)
}

// LogAndRespond logs an error and sends an appropriate response
func LogAndRespond(log *logger.Logger, w http.ResponseWriter, err error, statusCode int, correlationID string) {
	// Determine error type and message
	var code, message string
	switch statusCode {
	case http.StatusBadRequest:
		code = "BAD_REQUEST"
		message = "Invalid request"
	case http.StatusUnauthorized:
		code = "UNAUTHORIZED"
		message = "Authentication required"
	case http.StatusForbidden:
		code = "FORBIDDEN"
		message = "Access denied"
	case http.StatusNotFound:
		code = "NOT_FOUND"
		message = "Resource not found"
	case http.StatusConflict:
		code = "CONFLICT"
		message = "Resource conflict"
	case http.StatusTooManyRequests:
		code = "TOO_MANY_REQUESTS"
		message = "Rate limit exceeded"
	case http.StatusInternalServerError:
		code = "INTERNAL_ERROR"
		message = "Internal server error"
	case http.StatusServiceUnavailable:
		code = "SERVICE_UNAVAILABLE"
		message = "Service temporarily unavailable"
	default:
		code = "UNKNOWN_ERROR"
		message = "An unexpected error occurred"
	}

	// Log the error
	log.Error("Request failed",
		zap.String("error_code", code),
		zap.String("error_message", message),
		zap.Error(err),
		zap.String("correlation_id", correlationID),
	)

	// Send error response
	ErrorResponse(w, code, message, err.Error(), statusCode, correlationID)
}
