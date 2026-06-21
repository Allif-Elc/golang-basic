package utility

import (
	"errors"
	"net/http"
	"strings"
)

// Custom error types for better error handling
var (
	// ErrNotFound represents a resource not found error
	ErrNotFound = errors.New("not found")

	// ErrValidation represents a validation error
	ErrValidation = errors.New("validation failed")

	// ErrUnauthorized represents an unauthorized access error
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden represents a forbidden access error
	ErrForbidden = errors.New("forbidden")

	// ErrConflict represents a conflict error (e.g., duplicate resource)
	ErrConflict = errors.New("conflict")

	// ErrInternal represents an internal server error
	ErrInternal = errors.New("internal server error")
)

// AppError represents an application error with additional context
type AppError struct {
	Err       error
	Message   string
	Code      int
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

// Unwrap implements the errors.Unwrap interface
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(err error, message string, code int) *AppError {
	return &AppError{
		Err:     err,
		Message: message,
		Code:    code,
	}
}

// Error constructors for common scenarios
func NotFoundError(message string) *AppError {
	return &AppError{
		Err:     ErrNotFound,
		Message: message,
		Code:    http.StatusNotFound,
	}
}

func ValidationError(message string) *AppError {
	return &AppError{
		Err:     ErrValidation,
		Message: message,
		Code:    http.StatusBadRequest,
	}
}

func UnauthorizedError(message string) *AppError {
	return &AppError{
		Err:     ErrUnauthorized,
		Message: message,
		Code:    http.StatusUnauthorized,
	}
}

func ForbiddenError(message string) *AppError {
	return &AppError{
		Err:     ErrForbidden,
		Message: message,
		Code:    http.StatusForbidden,
	}
}

func ConflictError(message string) *AppError {
	return &AppError{
		Err:     ErrConflict,
		Message: message,
		Code:    http.StatusConflict,
	}
}

func InternalError(message string) *AppError {
	return &AppError{
		Err:     ErrInternal,
		Message: message,
		Code:    http.StatusInternalServerError,
	}
}

// GetHTTPStatusForError returns the appropriate HTTP status code for a given error
// It checks for AppError first, then falls back to error message analysis
func GetHTTPStatusForError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	// Check if it's an AppError
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}

	// Fallback to error message analysis
	errMsg := strings.ToLower(err.Error())

	// Check for unauthorized (must come before "not found" since "user id not found" contains both)
	if strings.Contains(errMsg, "unauthorized") ||
		strings.Contains(errMsg, "user id not found") {
		return http.StatusUnauthorized
	}

	// Check for not found
	if strings.Contains(errMsg, "not found") {
		return http.StatusNotFound
	}

	// Check for validation errors
	if strings.Contains(errMsg, "invalid") ||
		strings.Contains(errMsg, "required") ||
		strings.Contains(errMsg, "must") ||
		strings.Contains(errMsg, "too") ||
		strings.Contains(errMsg, "potentially dangerous") {
		return http.StatusBadRequest
	}

	// Check for forbidden
	if strings.Contains(errMsg, "forbidden") ||
		strings.Contains(errMsg, "permission") {
		return http.StatusForbidden
	}

	// Check for conflict
	if strings.Contains(errMsg, "already exists") ||
		strings.Contains(errMsg, "duplicate") {
		return http.StatusConflict
	}

	// Default to bad request
	return http.StatusBadRequest
}

// SendErrorResponse sends an error response with appropriate status code
func SendErrorResponse(w http.ResponseWriter, err error) {
	if err == nil {
		SendSuccess(w, http.StatusOK, "Success", nil)
		return
	}

	statusCode := GetHTTPStatusForError(err)
	SendError(w, statusCode, err.Error())
}

// IsNotFoundError checks if error is a not found error
func IsNotFoundError(err error) bool {
	if errors.Is(err, ErrNotFound) {
		return true
	}
	if strings.Contains(strings.ToLower(err.Error()), "not found") {
		return true
	}
	return false
}

// IsValidationError checks if error is a validation error
func IsValidationError(err error) bool {
	if errors.Is(err, ErrValidation) {
		return true
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "invalid") ||
		strings.Contains(errMsg, "required") ||
		strings.Contains(errMsg, "must") ||
		strings.Contains(errMsg, "too") ||
		strings.Contains(errMsg, "potentially dangerous")
}

// IsUnauthorizedError checks if error is an unauthorized error
func IsUnauthorizedError(err error) bool {
	if errors.Is(err, ErrUnauthorized) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "unauthorized")
}
