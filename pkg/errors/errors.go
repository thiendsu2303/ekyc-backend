package errors

import (
	"fmt"
	"net/http"
)

// Error represents a custom error with additional context
type Error struct {
	Code          int    `json:"code"`
	Message       string `json:"message"`
	Details       string `json:"details,omitempty"`
	RequestID     string `json:"request_id,omitempty"`
	SessionID     string `json:"session_id,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("code=%d, message=%s, details=%s", e.Code, e.Message, e.Details)
}

// New creates a new error with the given code and message
func New(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// WithDetails adds details to the error
func (e *Error) WithDetails(details string) *Error {
	e.Details = details
	return e
}

// WithRequestID adds request ID to the error
func (e *Error) WithRequestID(requestID string) *Error {
	e.RequestID = requestID
	return e
}

// WithSessionID adds session ID to the error
func (e *Error) WithSessionID(sessionID string) *Error {
	e.SessionID = sessionID
	return e
}

// WithCorrelationID adds correlation ID to the error
func (e *Error) WithCorrelationID(correlationID string) *Error {
	e.CorrelationID = correlationID
	return e
}

// Common errors
var (
	ErrInvalidInput       = New(http.StatusBadRequest, "Invalid input")
	ErrUnauthorized       = New(http.StatusUnauthorized, "Unauthorized")
	ErrForbidden          = New(http.StatusForbidden, "Forbidden")
	ErrNotFound           = New(http.StatusNotFound, "Resource not found")
	ErrConflict           = New(http.StatusConflict, "Resource conflict")
	ErrInternal           = New(http.StatusInternalServerError, "Internal server error")
	ErrServiceUnavailable = New(http.StatusServiceUnavailable, "Service unavailable")
	ErrTooManyRequests    = New(http.StatusTooManyRequests, "Too many requests")
)

// Database errors
var (
	ErrDatabaseConnection  = New(http.StatusInternalServerError, "Database connection failed")
	ErrDatabaseQuery       = New(http.StatusInternalServerError, "Database query failed")
	ErrDatabaseTransaction = New(http.StatusInternalServerError, "Database transaction failed")
	ErrRecordNotFound      = New(http.StatusNotFound, "Record not found")
	ErrDuplicateRecord     = New(http.StatusConflict, "Duplicate record")
)

// Validation errors
var (
	ErrInvalidEmail    = New(http.StatusBadRequest, "Invalid email format")
	ErrInvalidPassword = New(http.StatusBadRequest, "Invalid password")
	ErrInvalidID       = New(http.StatusBadRequest, "Invalid ID format")
	ErrMissingRequired = New(http.StatusBadRequest, "Missing required field")
	ErrInvalidFileType = New(http.StatusBadRequest, "Invalid file type")
	ErrFileTooLarge    = New(http.StatusBadRequest, "File too large")
)

// eKYC specific errors
var (
	ErrSessionExpired   = New(http.StatusUnauthorized, "Session expired")
	ErrDocumentRequired = New(http.StatusBadRequest, "Document required")
	ErrSelfieRequired   = New(http.StatusBadRequest, "Selfie required")
	ErrLivenessRequired = New(http.StatusBadRequest, "Liveness check required")
	ErrProcessingFailed = New(http.StatusInternalServerError, "Processing failed")
	ErrScoreTooLow      = New(http.StatusBadRequest, "Score too low")
	ErrAlreadyProcessed = New(http.StatusConflict, "Already processed")
)

// Storage errors
var (
	ErrFileUploadFailed    = New(http.StatusInternalServerError, "File upload failed")
	ErrFileNotFound        = New(http.StatusNotFound, "File not found")
	ErrFileDeleteFailed    = New(http.StatusInternalServerError, "File deletion failed")
	ErrInvalidPresignedURL = New(http.StatusBadRequest, "Invalid presigned URL")
)

// Event errors
var (
	ErrEventPublishFailed    = New(http.StatusInternalServerError, "Event publish failed")
	ErrEventSubscribeFailed  = New(http.StatusInternalServerError, "Event subscribe failed")
	ErrEventProcessingFailed = New(http.StatusInternalServerError, "Event processing failed")
)
