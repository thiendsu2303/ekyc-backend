package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ekyc-backend/pkg/logger"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// Validator holds the validation instance
type Validator struct {
	validate *validator.Validate
	logger   *logger.Logger
}

// NewValidator creates a new validator instance
func NewValidator(logger *logger.Logger) *Validator {
	v := validator.New()

	// Register custom validations if needed
	// v.RegisterValidation("custom_validation", customValidationFunc)

	return &Validator{
		validate: v,
		logger:   logger,
	}
}

// ValidateRequest validates the request body against a struct
func (v *Validator) ValidateRequest(w http.ResponseWriter, r *http.Request, target interface{}) bool {
	// Check content type
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		BadRequestResponse(w, "Invalid content type", "Content-Type must be application/json", "")
		return false
	}

	// Read and parse request body
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		v.logger.Error("Failed to decode request body",
			zap.Error(err),
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
		)
		BadRequestResponse(w, "Invalid JSON", "Request body must be valid JSON", "")
		return false
	}

	// Validate the struct
	if err := v.validate.Struct(target); err != nil {
		validationErrors := v.formatValidationErrors(err)
		v.logger.Warn("Request validation failed",
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.String("errors", validationErrors),
		)
		BadRequestResponse(w, "Validation failed", validationErrors, "")
		return false
	}

	return true
}

// ValidateQuery validates query parameters
func (v *Validator) ValidateQuery(w http.ResponseWriter, r *http.Request, target interface{}) bool {
	// Parse query parameters into the target struct
	if err := v.validate.Struct(target); err != nil {
		validationErrors := v.formatValidationErrors(err)
		v.logger.Warn("Query validation failed",
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.String("errors", validationErrors),
		)
		BadRequestResponse(w, "Invalid query parameters", validationErrors, "")
		return false
	}

	return true
}

// ValidatePath validates path parameters
func (v *Validator) ValidatePath(w http.ResponseWriter, r *http.Request, target interface{}) bool {
	// Parse path parameters into the target struct
	if err := v.validate.Struct(target); err != nil {
		validationErrors := v.formatValidationErrors(err)
		v.logger.Warn("Path validation failed",
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.String("errors", validationErrors),
		)
		BadRequestResponse(w, "Invalid path parameters", validationErrors, "")
		return false
	}

	return true
}

// formatValidationErrors formats validation errors into a readable string
func (v *Validator) formatValidationErrors(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errors []string
		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()
			param := e.Param()

			switch tag {
			case "required":
				errors = append(errors, field+" is required")
			case "email":
				errors = append(errors, field+" must be a valid email address")
			case "min":
				errors = append(errors, field+" must be at least "+param)
			case "max":
				errors = append(errors, field+" must be at most "+param)
			case "uuid":
				errors = append(errors, field+" must be a valid UUID")
			case "oneof":
				errors = append(errors, field+" must be one of: "+param)
			default:
				errors = append(errors, field+" failed validation: "+tag)
			}
		}
		return strings.Join(errors, "; ")
	}
	return err.Error()
}

// ValidateStruct validates a struct directly
func (v *Validator) ValidateStruct(s interface{}) error {
	return v.validate.Struct(s)
}

// ValidateField validates a single field
func (v *Validator) ValidateField(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}

// RegisterCustomValidation registers a custom validation function
func (v *Validator) RegisterCustomValidation(tag string, fn validator.Func) error {
	return v.validate.RegisterValidation(tag, fn)
}

// GetValidator returns the underlying validator instance
func (v *Validator) GetValidator() *validator.Validate {
	return v.validate
}
